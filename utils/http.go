package utils

import (
	"net"
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		// Modified from https://go.dev/src/net/http/transport.go:DefaultTransport
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			ForceAttemptHTTP2:     false, // for more easy debugging
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
