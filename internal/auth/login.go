package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/secrets"
	"unipilot/internal/services/database"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
)

// Login authenticates a user with the provided credentials and initializes the session.
// Performs authentication, saves tokens, creates authenticated HTTP client, initializes SSE connection,
// and triggers post-login data migration (courses and assignments).
//
// Parameters:
//   - username: User's username for authentication
//   - password: User's password in plain text
//
// Returns:
//   - *models.User: Authenticated user object with profile information
//   - error: Error if authentication fails, token saving fails, or post-login operations fail
func Login(username, password string) (*database.Database, *Auth, error) {
	// Step 1: Create HTTP client with cookie jar for session management
	// Cookie jar stores authentication cookies automatically
	httpClient := http.Client{
		Jar: &client.CookieJar{},
	}

	// Step 2: Prepare login request payload
	loginData := map[string]string{"username": username, "password": password}
	jsonData, _ := json.Marshal(loginData)

	api_url := secrets.CONSTANTS["API_URL"]

	// Step 3: Send authentication request to API
	resp, err := httpClient.Post(fmt.Sprintf("%s/auth/login", api_url), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.NetworkConnectionFailed, "HTTP POST failed")
	}
	defer resp.Body.Close()

	// Step 4: Validate response status code
	if resp.StatusCode != http.StatusOK {
		var serverError *errors.ServerError
		if err := json.NewDecoder(resp.Body).Decode(&serverError); err != nil {
			return nil, nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse login response")
		}
		return nil, nil, serverError
	}

	// Step 5: Parse API response to extract user data and tokens
	var response struct {
		User         map[string]interface{} `json:"user"`
		Token        string                 `json:"token"`
		RefreshToken string                 `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
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

	// Step 7: Persist user credentials to local storage for future sessions
	if err := utils.SetCredentials(&response_user); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to set credentials")
	}

	// Step 8: Save JWT access token for authenticated API requests
	if err := client.SaveToken(response.Token); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	// Step 9: Save refresh token for token renewal without re-authentication
	if err := client.SaveRefreshToken(response.RefreshToken); err != nil {
		return nil, nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
	}

	return PostLogin(&response_user)
}

// PostLogin performs post-authentication data migration tasks.
// Initializes local database schema and migrates courses and assignments from remote server.
// All operations are non-fatal - errors are logged but don't prevent successful login.
//
// Returns:
//   - error: Last error encountered (if any), but login succeeds regardless
func PostLogin(user *models.User) (*database.Database, *Auth, error) {

	httpClient, err := client.NewAuthClient()
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ClientRequestFailed, "Failed to create authenticated HTTP client")
	}

	a := &Auth{
		User:   user,
		SSE:    sse.NewSSE(),
		Client: httpClient,
	}

	database, err := database.NewDatabase(a.User)
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.DBConnectionFailed, "Failed to initialize database")
	}

	return database, a, nil
}
