package client

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// getAuthToken loads and refreshes token if needed, returns the token string
func getAuthToken() (string, error) {
	token, err := LoadToken()
	if err != nil {
		return "", fmt.Errorf("failed to load token: %w", err)
	}

	// Refresh token if it is about to expire
	if !IsTokenValid() {
		refreshToken, err := LoadRefreshToken()
		if err != nil {
			return "", fmt.Errorf("failed to load refresh token: %w", err)
		}

		newToken, newRefreshToken, err := RefreshToken(refreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}

		token = newToken
		if err := SaveToken(newToken); err != nil {
			return "", fmt.Errorf("failed to save token: %w", err)
		}
		if err := SaveRefreshToken(newRefreshToken); err != nil {
			return "", fmt.Errorf("failed to save refresh token: %w", err)
		}
		log.Println("Token refreshed")
	}

	return token, nil
}

// setAuthHeader sets the Authorization header on a fiber agent with automatic token refresh
func setAuthHeader(agent *fiber.Agent) error {
	token, err := getAuthToken()
	if err != nil {
		return err
	}

	// Add Authorization header if token exists
	if token != "" {
		agent.Set("Authorization", "Bearer "+token)
	}

	return nil
}

// setAuthHeaderRequest sets the Authorization header on an http.Request with automatic token refresh
func setAuthHeaderRequest(req *http.Request) error {
	token, err := getAuthToken()
	if err != nil {
		return err
	}

	// Add Authorization header if token exists
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return nil
}
