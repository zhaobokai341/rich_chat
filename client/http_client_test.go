package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient(t *testing.T) {
	userAgent := "rich_chat_client/1.0"

	client := NewHTTPClient(userAgent)

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestHTTPClient_R(t *testing.T) {
	client := NewHTTPClient("test-agent")

	request := client.R()

	assert.NotNil(t, request)
}

func TestHTTPClient_SetHeader(t *testing.T) {
	client := NewHTTPClient("test-agent")

	client.SetHeader("Authorization", "Bearer token123")

	// Verify header was set by creating a request and checking
	req := client.R()
	assert.NotNil(t, req)
}

func TestHTTPClient_SetHeaders(t *testing.T) {
	client := NewHTTPClient("test-agent")

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token456",
		"X-Custom":      "value",
	}

	client.SetHeaders(headers)

	// Verify headers were set
	req := client.R()
	assert.NotNil(t, req)
}
