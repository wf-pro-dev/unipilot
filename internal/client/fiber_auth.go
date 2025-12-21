package client

import (
	"net/http"

	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

// getAuthToken loads and refreshes token if needed, returns the token string
func getAuthToken() (string, error) {
	token, err := LoadToken()
	if err != nil {
		return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to load token")
	}

	// Refresh token if it is about to expire
	if !IsTokenValid() {
		refreshToken, err := LoadRefreshToken()
		if err != nil {
			return "", errors.Wrap(err, errors.FSFileNotFound, "Failed to load refresh token")
		}

		newToken, newRefreshToken, err := RefreshToken(refreshToken)
		if err != nil {
			return "", errors.Wrap(err, errors.ClientRequestFailed, "Failed to refresh token")
		}

		token = newToken
		if err := SaveToken(newToken); err != nil {
			return "", errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
		}
		if err := SaveRefreshToken(newRefreshToken); err != nil {
			return "", errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
		}
	}

	return token, nil
}

// SetAuthHeader sets the Authorization header on a fiber agent with automatic token refresh
func SetAuthHeader(agent *fiber.Agent) error {
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

// SetAuthHeaderRequest sets the Authorization header on an http.Request with automatic token refresh
func SetAuthHeaderRequest(req *http.Request) error {
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
