package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
	"unipilot/internal/sync"
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
//   - *user.User: Authenticated user object with profile information
//   - error: Error if authentication fails, token saving fails, or post-login operations fail
func (a *Auth) Login(username, password string) (*user.User, error) {
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
		return nil, errors.Wrap(err, errors.NetworkConnectionFailed, "HTTP POST failed")
	}
	defer resp.Body.Close()

	// Step 4: Validate response status code
	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewAppError(errors.AuthUnauthorized, "Login failed", nil).ToServerError(resp.StatusCode)
	}

	// Step 5: Parse API response to extract user data and tokens
	var response struct {
		User         map[string]interface{} `json:"user"`
		Token        string                 `json:"token"`
		RefreshToken string                 `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse login response")
	}

	// Step 6: Convert response map to User struct
	// Type assertions extract string values from interface{} map
	response_user := user.User{
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
		return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to set credentials")
	}

	// Step 8: Save JWT access token for authenticated API requests
	if err := client.SaveToken(response.Token); err != nil {
		return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save token")
	}

	// Step 9: Save refresh token for token renewal without re-authentication
	if err := client.SaveRefreshToken(response.RefreshToken); err != nil {
		return nil, errors.Wrap(err, errors.FSWriteFailed, "Failed to save refresh token")
	}

	// Step 10: Create authenticated HTTP client with saved tokens
	httpUserClient, err := client.NewAuthClient()
	if err != nil {
		return nil, errors.Wrap(err, errors.ClientRequestFailed, "Could not create authenticated HTTP client")
	}

	a.Client = httpUserClient

	// Step 11: Initialize SSE connection for real-time notifications
	// Initialize early to ensure it's never nil for subsequent operations
	a.SSE = sse.NewSSE()

	// Step 12: Perform post-login data migration (courses and assignments)
	// Errors are non-fatal - login succeeds even if migration fails
	if err := PostLogin(); err != nil {
		return &response_user, err
	}

	return &response_user, nil
}

// PostLogin performs post-authentication data migration tasks.
// Initializes local database schema and migrates courses and assignments from remote server.
// All operations are non-fatal - errors are logged but don't prevent successful login.
//
// Returns:
//   - error: Last error encountered (if any), but login succeeds regardless
func PostLogin() error {

	// Step 1: Get local database connection for data migration
	// GetUserDB will create the database file if it doesn't exist
	localDB, err := utils.GetUserDB()
	if err != nil {
		// If we can't get the local database, log and return the error
		fmt.Printf("Warning: Could not get local database: %v\n", err)
		fmt.Printf("Login successful, but database operations failed\n")
		return err
	}

	log.Println("Local database connected", localDB)

	// Step 2: Initialize database schema (create tables if they don't exist)
	// Non-fatal operation - login succeeds even if schema initialization fails
	if err := storage.InitializeSchema(localDB); err != nil {
		log.Println("Failed to initialize database schema", err)
		// Don't fail the login, just continue
	}

	// Step 3: Migrate courses from remote server to local database
	// Non-fatal operation - allows offline access to courses
	if err := sync.MigrateCourses(localDB); err != nil {
		log.Println("Failed to migrate courses", err)
		// Don't rollback, continue with the transaction
	}

	// Step 4: Migrate assignments from remote server to local database
	// Non-fatal operation - allows offline access to assignments
	if err := sync.MigrateAssignments(localDB); err != nil {
		log.Println("Failed to migrate assignments", err)
		// Don't rollback, continue with the transaction
	}

	return nil
}
