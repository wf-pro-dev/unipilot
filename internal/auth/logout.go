package auth

import (
	"fmt"
	"net/http"

	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
)

// Logout terminates the user session and clears local authentication data.
// Sends logout request to server, clears local tokens and credentials.
// Note: With stateless JWT tokens, server-side logout is primarily for audit logging.
//
// Returns:
//   - error: Error if logout request fails or local cleanup fails
func (a *Auth) Logout() error {
	// Step 1: Clear user from memory immediately
	// Prevents any further authenticated operations
	a.User = nil

	api_url := secrets.CONSTANTS["API_URL"]

	// Step 2: Send logout request to server for audit logging
	// With stateless JWT tokens, this is primarily for server-side logging
	// Make POST request to logout endpoint (empty body)
	resp, err := a.Client.Post(
		fmt.Sprintf("%s/auth/logout", api_url),
		"application/json",
		nil,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Step 3: Validate server response (200 OK indicates successful logout acknowledgment)
	// Only consider status 200 OK as successful logout
	if resp.StatusCode != http.StatusOK {
		return errors.NewAppError(errors.AuthUnauthorized, "Logout failed", nil).ToServerError(resp.StatusCode)
	}

	// Step 4: Clear local cookies regardless of server response
	// Ensures complete local cleanup even if server request fails
	if err := client.ClearCookies(); err != nil {
		return errors.Wrap(err, errors.FSDeleteFailed, "Failed to clear local cookies")
	}

	// Step 5: Clear saved user credentials from local storage
	if err := utils.ClearCredentials(); err != nil {
		return errors.Wrap(err, errors.FSDeleteFailed, "Failed to clear local credentials")
	}

	return nil
}
