package main

import (
	"github.com/go-resty/resty/v2"
)

// HTTPClient wraps resty.Client to provide a testable interface
type HTTPClient struct {
	client *resty.Client
}

// NewHTTPClient creates a new HTTP client with default configuration
func NewHTTPClient(userAgent string) *HTTPClient {
	client := resty.New()
	client.SetHeader("User-Agent", userAgent)
	return &HTTPClient{
		client: client,
	}
}

// R returns a new request instance
func (h *HTTPClient) R() *resty.Request {
	return h.client.R()
}

// SetHeader sets a header for all requests
func (h *HTTPClient) SetHeader(key, value string) {
	h.client.SetHeader(key, value)
}

// SetHeaders sets multiple headers for all requests
func (h *HTTPClient) SetHeaders(headers map[string]string) {
	h.client.SetHeaders(headers)
}
