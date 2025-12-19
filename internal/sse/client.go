package sse

import (
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
)

// NewSSEClientWithJWT creates an HTTP client for SSE with JWT authentication
func NewSSEClient() (*http.Client, error) {

	httpClient, err := client.NewAuthClient()
	if err != nil {
		return nil, errors.Wrap(err, errors.ClientRequestFailed, "Failed to create auth client")
	}

	// Configure transport specifically for SSE connections
	transport := &http.Transport{
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second, // Increased from 10s
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     false,
		// Add these for better connection handling
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient.Transport = &client.JWTRoundTripper{
		Base:  transport,
		Token: "", // Will be loaded on first request
	}

	httpClient.Timeout = 0

	return httpClient, nil
}
