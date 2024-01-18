package httptransmit

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/NIR3X/fxms"
)

func request(urlAddr, method string, headers []string, data []uint8, httpHeaders http.Header) (int, []uint8) {
	req, err := http.NewRequest(method, urlAddr, bytes.NewBuffer(data))
	if err != nil {
		return 0, nil
	}
	for _, header := range headers {
		index := strings.Index(header, ": ")
		if index >= 0 {
			req.Header.Set(header[:index], header[index+2:])
		}
	}

	req.Header.Set("X-Forwarded-For", strings.Trim(httpHeaders.Get("Cf-Connecting-Ip")+","+httpHeaders.Get("X-Forwarded-For"), ","))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil
	}

	return resp.StatusCode, respData
}

type HttpTransmitSession struct {
	key        []uint8
	lastUpdate int64
}

func NewHttpTransmitSession(key []uint8) *HttpTransmitSession {
	return &HttpTransmitSession{
		key:        key,
		lastUpdate: time.Now().Unix(),
	}
}

func (httpTransmitSession *HttpTransmitSession) updateLastUpdate() {
	httpTransmitSession.lastUpdate = time.Now().Unix()
}

func (httpTransmitSession *HttpTransmitSession) getDeltaUpdate() int64 {
	return time.Now().Unix() - httpTransmitSession.lastUpdate
}

type HttpTransmit struct {
	mtx                sync.RWMutex
	stopChan           chan struct{}
	whitelistedHosts   map[string]bool
	key                []uint8
	maxSessionTimeSecs int64
	sessions           map[string]*HttpTransmitSession
}

func NewHttpTransmit(whitelistedHosts map[string]bool, key [fxms.KeyLen]uint8, maxSessionTimeSecs int64) (*HttpTransmit, error) {
	if len(key) != fxms.KeyLen {
		return nil, errors.New("Invalid key length.")
	}

	httpTransmit := &HttpTransmit{
		mtx:                sync.RWMutex{},
		stopChan:           make(chan struct{}),
		whitelistedHosts:   whitelistedHosts,
		key:                key[:],
		maxSessionTimeSecs: maxSessionTimeSecs,
		sessions:           make(map[string]*HttpTransmitSession),
	}

	go func() {
		for {
			select {
			case <-httpTransmit.stopChan:
				return
			default:
				httpTransmit.mtx.Lock()
				maxSessionTimeSecs := httpTransmit.maxSessionTimeSecs
				for sessionId, session := range httpTransmit.sessions {
					if session.getDeltaUpdate() >= maxSessionTimeSecs {
						delete(httpTransmit.sessions, sessionId)
					}
				}
				httpTransmit.mtx.Unlock()
				time.Sleep(time.Duration(maxSessionTimeSecs) * time.Second)
			}
		}
	}()

	return httpTransmit, nil
}

func (httpTransmit *HttpTransmit) Close() {
	httpTransmit.stopChan <- struct{}{}
}

func (httpTransmit *HttpTransmit) HandleConnect(w http.ResponseWriter, r *http.Request) {
	sessionKeyEncB64 := r.Header.Get("HT-Session-Key")
	sessionKeyEnc, err := base64.StdEncoding.DecodeString(sessionKeyEncB64)
	if err != nil {
		return
	}

	if len(sessionKeyEnc) < fxms.HashLen+fxms.MaskLen+fxms.KeyLen {
		return
	}

	sessionKey, ok, _ := fxms.Decrypt(httpTransmit.key, sessionKeyEnc, fxms.OptimizeDecryption)
	if !ok {
		return
	}

	if len(sessionKey) < fxms.KeyLen {
		return
	}

	sessionId := r.Header.Get("HT-Session-ID")

	httpTransmit.mtx.Lock()
	if _, ok := httpTransmit.sessions[sessionId]; !ok {
		httpTransmit.sessions[sessionId] = NewHttpTransmitSession(sessionKey)
	}
	httpTransmit.mtx.Unlock()

	if data, err := fxms.Encrypt(sessionKey, []uint8{}, fxms.OptimizeEncryption); err == nil {
		w.Write(data)
	}
}

func (httpTransmit *HttpTransmit) HandleTransmit(w http.ResponseWriter, r *http.Request) {
	sessionId := r.Header.Get("HT-Session-ID")
	httpTransmit.mtx.Lock()
	session, ok := httpTransmit.sessions[sessionId]
	if ok {
		session.updateLastUpdate()
	}
	httpTransmit.mtx.Unlock()

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	headersEncB64 := r.Header.Get("HT-Session-Headers")
	headersEnc, err := base64.StdEncoding.DecodeString(headersEncB64)
	if err != nil {
		return
	}

	if len(headersEnc) < fxms.HashLen+fxms.MaskLen {
		return
	}

	headersDec, ok, _ := fxms.Decrypt(session.key, headersEnc, fxms.OptimizeDecryption)
	if !ok {
		return
	}

	headersLines := strings.Split(string(headersDec), "\n")
	if len(headersLines) < 2 {
		return
	}

	urlAddr := headersLines[0]
	method := headersLines[1]
	headers := headersLines[2:]

	urlParsed, err := url.Parse(urlAddr)
	if err != nil {
		return
	}

	httpTransmit.mtx.RLock()
	_, ok = httpTransmit.whitelistedHosts[urlParsed.Host]
	httpTransmit.mtx.RUnlock()

	if !ok {
		return
	}

	dataEnc, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	data, ok, _ := fxms.Decrypt(session.key, dataEnc, fxms.OptimizeDecryption)
	if !ok {
		return
	}

	respStatusCode, respData := request(urlAddr, method, headers, data, r.Header)
	w.WriteHeader(respStatusCode)
	if data, err := fxms.Encrypt(session.key, respData, fxms.OptimizeEncryption); err == nil {
		w.Write(data)
	}
}
