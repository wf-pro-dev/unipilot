package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unipilot/internal/secrets"
)

// GetUser retrieves the current authenticated user's profile from the API.
// Uses the authenticated HTTP client to fetch user data from the server.
// This is useful for refreshing user information or verifying authentication status.
//
// Returns:
//   - map[string]interface{}: User data as a map (username, email, avatar, etc.)
//   - error: Error if request fails, authentication fails, or response parsing fails
func (a *Auth) GetUser() (map[string]interface{}, error) {
	api_url := secrets.CONSTANTS["API_URL"]

	// Step 1: Send GET request to user endpoint using authenticated client
	// Client includes JWT token in Authorization header automatically
	resp, err := a.Client.Get(fmt.Sprintf("%s/user", api_url))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Step 2: Validate response status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	// Step 3: Parse JSON response
	var response struct {
		Message string                 `json:"message"`
		User    map[string]interface{} `json:"user"`
		Error   string                 `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Step 4: Check for error in response body
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}

	// Step 5: Validate user data exists in response
	if response.User == nil {
		return nil, fmt.Errorf("no user data in response")
	}

	return response.User, nil
}
