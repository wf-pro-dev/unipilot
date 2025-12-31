package auth

import (
	"log"
	"net/http"
	"unipilot/internal/client"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
)

// Auth represents the authentication context for the application.
// It manages the authenticated HTTP client, SSE connection, and current user state.
type Auth struct {
	Client *http.Client // Authenticated HTTP client for API requests
	SSE    *sse.SSE     // Server-Sent Events connection for real-time updates
	User   *models.User   // Current authenticated user
}

// NewAuth creates and initializes a new Auth instance.
// Attempts to restore authentication state from local storage if available.
// If a user is found in local storage, creates an authenticated HTTP client.
//
// Returns:
//   - *Auth: Initialized Auth instance (may be unauthenticated if no saved credentials exist)
func NewAuth() *Auth {
	// Step 1: Initialize empty Auth struct with nil values
	newAuth := &Auth{
		Client: nil,
		SSE:    nil,
		User:   nil,
	}

	// Step 2: Attempt to restore user from local credentials file
	// This allows the app to remember authentication across restarts
	currentUser, err := utils.GetUserFromFile()
	if err == nil {
		newAuth.User = currentUser

		// Step 3: Create authenticated HTTP client using saved credentials
		// This client includes JWT tokens for API requests
		newAuth.Client, err = client.NewAuthClient()
		if err != nil {
			log.Printf("failed to create client: %v", err)
		}

	}

	return newAuth
}

// IsAuthenticated checks if the Auth instance has an authenticated models.
//
// Returns:
//   - bool: true if user is authenticated, false otherwise
func (a *Auth) IsAuthenticated() bool {
	return a.User != nil
}
