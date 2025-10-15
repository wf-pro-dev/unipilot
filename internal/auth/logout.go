package auth

import (
	"fmt"
	"log"
	"net/http"

	"unipilot/internal/client"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
)

func (a *Auth) Logout() error {

	a.User = nil

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return fmt.Errorf("failed to get api url: %w", err)
	}

	// Make POST request to logout endpoint (empty body)
	resp, err := a.Client.Post(
		fmt.Sprintf("%s/logout", api_url),
		"application/json",
		nil,
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

	log.Println("Logout: Clearing credentials")
	if err := utils.ClearCredentials(); err != nil {
		return fmt.Errorf("failed to clear local credentials: %w", err)
	}

	return nil
}
