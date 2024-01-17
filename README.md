# HTTP Transmit - Secure HTTP Proxy with FXMS Symmetric Encryption

HTTP Transmit is a Go-based proxy server that facilitates encrypted proxying of HTTP requests using the FXMS symmetric encryption algorithm. It includes features for session management, encryption, and automatic cleanup of expired sessions. This proxy server is designed for secure communication between clients and destination servers.

## Features

* **Session Management:** Manages individual sessions, including session keys and last update times.
* **Encryption and Decryption:** Utilizes the FXMS package for secure encryption and decryption of session keys, headers, and data.
* **Whitelisting:** Supports whitelisting of specific hosts to ensure that requests are only proxied for authorized destinations.
* **Automatic Session Cleanup:** Periodically cleans up expired sessions to prevent memory leaks.
* **HTTP Connect Handler (`HandleConnect`):** Establishes new sessions and responds with encrypted session keys.
* **HTTP Transmit Handler (`HandleTransmit`):** Transmits encrypted data to the destination server and forwards the encrypted response back to the client.

## Installation

To use this package, you can install it using go get command:

```bash
go get github.com/NIR3X/httptransmit
```

# Usage

Here is an example of how to use this package:

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/NIR3X/httptransmit"
)

func main() {
	whitelistedHosts := map[string]bool{
		"127.0.0.1:50001": true,
		"127.0.0.1:50120": true,
		"127.0.0.1:52001": true,
		"127.0.0.1:52002": true,
	}

	key := [256]uint8{ /* ... */ }

	httpTransmit, err := httptransmit.NewHttpTransmit(whitelistedHosts, key, 60 /* maxSessionTimeSecs */)
	if err != nil {
		panic(err)
	}
	defer httpTransmit.Close()

	http.HandleFunc("/connect", httpTransmit.HandleConnect)
	http.HandleFunc("/transmit", httpTransmit.HandleTransmit)

	log.Println("Server is started")

	if err := http.ListenAndServe(":53440", nil); err != nil {
		panic(err)
	}
}
```
