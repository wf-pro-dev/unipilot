package auth

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"
	"unipilot/internal/services/database"
	"unipilot/internal/services/utils"
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
func (a *Auth) Register(userData *models.User) (*database.Database, *Auth, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/auth/register", api_url))
	agent.JSON(userData)

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, nil, errs[0]
	}

	if statusCode != 200 {
		serverError := errors.ParseServerError(body, statusCode)
		return nil, nil, serverError
	}

	var response struct {
		User         models.User `json:"user"`
		Token        string      `json:"token"`
		RefreshToken string      `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	// Step 7: Persist user credentials to local storage
	if err := utils.SetCredentials(&response.User); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to set credentials")
	}

	// Step 8: Save JWT access token for authenticated API requests
	if err := client.SaveToken(response.Token); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	if err := client.SaveRefreshToken(response.RefreshToken); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
	}

	dbService, authService, err := PostLogin(&response.User)
	if err != nil {
		return nil, nil, err
	}
	return dbService, authService, nil
}
