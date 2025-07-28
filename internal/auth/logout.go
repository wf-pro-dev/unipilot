package auth

import (
	"fmt"
	"net/http"

	"unipilot/internal/client"
	"unipilot/internal/storage"
)

func (a *Auth) Logout() error {

	// Make POST request to logout endpoint (empty body)
	resp, err := a.Client.Post(
		"https://newsroom.dedyn.io/acc-homework/logout", // Note: changed from /login to /logout
		"application/json",
		nil, // No body needed for logout
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Only consider status 200 OK as successful logout
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
	}

	// Clear local cookies regardless of server response
	if err := client.ClearCookies(); err != nil {
		return fmt.Errorf("failed to clear local cookies: %w", err)
	}

	if err := storage.ClearCredentials(); err != nil {
		return fmt.Errorf("failed to clear local credentials: %w", err)
	}

	return nil
}
