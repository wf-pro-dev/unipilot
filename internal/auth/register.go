package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	// Step 5: Parse API response to extract user data and tokens
	var response struct {
		User         map[string]interface{} `json:"user"`
		Token        string                 `json:"token"`
		RefreshToken string                 `json:"refresh_token"`
	}
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&response); err != nil {
		return nil, nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse login response")
	}

	// Step 6: Convert response map to User struct
	// Type assertions extract string values from interface{} map
	response_user := models.User{
		Username:   response.User["username"].(string),
		Email:      response.User["email"].(string),
		Avatar:     response.User["avatar"].(string),
		University: response.User["university"].(string),
		Semester:   response.User["semester"].(string),
		Year:       response.User["year"].(string),
		Language:   response.User["language"].(string),
	}

	// Parse timestamps from RFC3339 format
	response_user.CreatedAt, _ = time.Parse(time.RFC3339, response.User["created_at"].(string))
	response_user.UpdatedAt, _ = time.Parse(time.RFC3339, response.User["updated_at"].(string))

	// Convert float64 ID to uint (JSON numbers are float64)
	response_user.ID = uint(response.User["id"].(float64))

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
	}

	log.Println("response_user: ", response_user)
	// Step 7: Persist user credentials to local storage
	if err := utils.SetCredentials(&response_user); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to set credentials")
	}

	// Step 8: Save JWT access token for authenticated API requests
	if err := client.SaveToken(response.Token); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	if err := client.SaveRefreshToken(response.RefreshToken); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
	}

	dbService, authService, err := PostLogin(&response_user)
	if err != nil {
		return nil, nil, err
	}
	return dbService, authService, nil
}
