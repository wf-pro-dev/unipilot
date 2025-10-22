package client

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// NewAuthClient creates a client with JWT token authentication
func NewAuthClient() (*http.Client, error) {
	// Configure transport
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	// Create client with custom transport
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Add JWT token to all requests
	client.Transport = &JWTRoundTripper{
		Base:  transport,
		Token: "", // Will be loaded on first request
	}

	return client, nil
}

// JWTRoundTripper adds JWT token to requests
type JWTRoundTripper struct {
	Base         http.RoundTripper
	Token        string
	RefreshToken string
}

func (rt *JWTRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Load token if not already loaded
	if rt.Token == "" {
		token, err := LoadToken()
		if err == nil && token != "" {
			rt.Token = token
		}
	}

	if rt.RefreshToken == "" {
		refreshToken, err := LoadRefreshToken()
		if err == nil && refreshToken != "" {
			rt.RefreshToken = refreshToken
		}
	}

	// Refresh token if it is about to expire
	if !IsTokenValid() {

		// call /refresh-token endpoint to get a new token
		newToken, newRefreshToken, err := RefreshToken(rt.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		rt.Token = newToken
		if err := SaveToken(newToken); err != nil {
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
		rt.RefreshToken = newRefreshToken
		if err := SaveRefreshToken(newRefreshToken); err != nil {
			return nil, fmt.Errorf("failed to save refresh token: %w", err)
		}
		log.Println("Token refreshed")

	}

	// Add Authorization header if token exists
	if rt.Token != "" {
		req.Header.Set("Authorization", "Bearer "+rt.Token)
	}

	return rt.Base.RoundTrip(req)
}
