package client

import (
	"net/http"
	"time"
)

type HTTPClient struct {
	Client  *http.Client
	BaseURL string
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		Client:  &http.Client{Timeout: timeout},
		BaseURL: baseURL,
	}
}
