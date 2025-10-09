package httpclient

import (
	"net"
	"net/http"
	"time"
)

func New() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
	}
}
