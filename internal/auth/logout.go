package auth

import (
	"fmt"
	"log"
	"net/http"

	"unipilot/internal/client"
	"unipilot/internal/services/utils"
)

func (a *Auth) Logout() error {

	a.User = nil

	log.Printf("Logout: %v", a.Client)
	// Make POST request to logout endpoint (empty body)
	resp, err := a.Client.Post(
		"https://newsroom.dedyn.io/acc-homework/logout",
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
