package auth

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
)

// Register creates a new user account and automatically authenticates the models.
// Performs user registration, saves tokens, initializes SSE connection, and sets up local database.
// Similar to Login but creates a new account instead of authenticating an existing one.
//
// Parameters:
//   - username: Desired username (must be unique)
//   - email: User's email address
//   - password: User's password in plain text (will be hashed server-side)
//   - university: User's university affiliation
//   - language: User's preferred language
//
// Returns:
//   - *models.User: Newly created user object with profile information
//   - error: Error if registration fails, token saving fails, or database setup fails
func (a *Auth) Register(userData *models.User) (*models.User, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/auth/register", api_url))
	agent.JSON(userData)

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, serverError
	}

	var response struct {
		User         models.User `json:"user"`
		Token        string      `json:"token"`
		RefreshToken string      `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	// Step 7: Persist user credentials to local storage
	if err := utils.SetCredentials(&response.User); err != nil {
		return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to set credentials")
	}

	// Step 8: Save JWT access token for authenticated API requests
	if err := client.SaveToken(response.Token); err != nil {
		return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	// Step 9: Initialize SSE connection for real-time notifications
	// Initialize early to ensure it's never nil for subsequent operations
	a.SSE = sse.NewSSE()

	// Step 10: Initialize local database schema for offline data storage
	// Gracefully handle missing database directory (first-time registration scenario)
	localDB, err := utils.GetUserDB()
	if err != nil {
		// If we can't get the local database, just log it and continue
		// This might happen if the database directory doesn't exist yet
		fmt.Printf("Warning: Could not get local database: %v\n", err)
		fmt.Printf("Login successful, but database operations failed\n")
		return &response.User, nil // Don't fail the registration, just return success
	}
	// Initialize the database schema (create tables if they don't exist)
	// Non-fatal operation - registration succeeds even if schema initialization fails
	if err := storage.InitializeSchema(localDB); err != nil {
		fmt.Printf("Warning: Failed to initialize database schema: %v\n", err)
		// Don't fail the registration, just continue
	}

	return &response.User, nil
}
