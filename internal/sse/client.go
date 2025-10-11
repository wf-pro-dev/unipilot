package sse

import (
	"net/http"
	"time"
	"unipilot/internal/client"
)

// NewSSEClientWithJWT creates an HTTP client for SSE with JWT authentication
func NewSSEClient() (*http.Client, error) {
	// Configure transport specifically for SSE connections
	transport := &http.Transport{
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     false,
	}

	// Create client with JWT authentication
	client := &http.Client{
		Transport: &client.JWTRoundTripper{
			Base:  transport,
			Token: "", // Will be loaded on first request
		},

		Timeout: 0, // No timeout for SSE connections
	}

	return client, nil
}
