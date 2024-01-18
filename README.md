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
go get -u github.com/NIR3X/httptransmit
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
		"127.0.0.1:50002": true,
		"127.0.0.1:50003": true,
		"example.com": true,
	}

	key := [256]uint8{
		183, 100, 137, 132, 244, 170, 22, 121, 90, 211, 232, 159, 161, 191, 191, 204,
		90, 151, 91, 93, 160, 170, 49, 24, 217, 40, 144, 53, 158, 135, 18, 155,
		53, 246, 43, 92, 16, 194, 241, 104, 81, 173, 18, 60, 202, 244, 194, 252,
		173, 58, 168, 3, 125, 142, 188, 238, 14, 76, 215, 179, 251, 127, 129, 129,
		86, 230, 199, 145, 65, 9, 141, 11, 235, 192, 237, 167, 207, 93, 234, 98,
		229, 239, 163, 105, 138, 151, 77, 3, 45, 79, 181, 162, 157, 38, 176, 7,
		163, 172, 160, 55, 3, 57, 149, 148, 54, 91, 54, 87, 192, 191, 62, 100,
		176, 215, 90, 229, 110, 197, 103, 166, 224, 32, 212, 115, 32, 189, 128, 1,
		27, 96, 170, 0, 154, 229, 207, 62, 117, 165, 69, 72, 20, 162, 41, 76,
		235, 93, 70, 18, 1, 99, 48, 134, 52, 51, 176, 178, 20, 251, 168, 211,
		12, 181, 65, 102, 190, 103, 73, 11, 224, 221, 115, 48, 144, 236, 206, 171,
		95, 2, 222, 207, 8, 57, 165, 202, 102, 73, 86, 67, 62, 239, 89, 212,
		237, 44, 216, 121, 161, 38, 82, 65, 247, 130, 133, 52, 234, 162, 167, 191,
		109, 16, 239, 99, 89, 18, 156, 211, 77, 179, 94, 73, 97, 175, 1, 39,
		68, 81, 101, 217, 117, 82, 220, 181, 177, 120, 109, 2, 107, 208, 74, 228,
		242, 188, 34, 174, 33, 107, 184, 237, 200, 153, 41, 13, 131, 80, 234, 202,
	}

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
