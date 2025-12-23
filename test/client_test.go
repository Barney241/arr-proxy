package test

import "net/http"

func newTestClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
	}
}