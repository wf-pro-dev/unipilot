package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
)

// Register creates a new user account and automatically authenticates the user.
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
//   - *user.User: Newly created user object with profile information
//   - error: Error if registration fails, token saving fails, or database setup fails
func (a *Auth) Register(username, email, password, university, language string) (*user.User, error) {
	// Step 1: Create HTTP client (unauthenticated for registration)
	// Registration endpoint doesn't require authentication
	httpClient, err := client.NewAuthClient()
	if err != nil {
		return nil, errors.Wrap(err, errors.ClientRequestFailed, "Could not create HTTP client")
	}

	// Step 2: Prepare registration request payload
	loginData := map[string]string{"username": username, "password": password, "email": email, "university": university, "language": language}
	jsonData, _ := json.Marshal(loginData)

	api_url := secrets.CONSTANTS["API_URL"]

	// Step 3: Send registration request to API
	resp, err := httpClient.Post(fmt.Sprintf("%s/register", api_url), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, errors.NetworkConnectionFailed, "HTTP POST failed")
	}
	defer resp.Body.Close()

	fmt.Println("Register response status code: ", resp.StatusCode)

	// Step 4: Validate response status (accept both 200 OK and 201 Created)
	// 201 Created is standard for resource creation, but 200 OK is also acceptable
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		io.ReadAll(resp.Body) // Read body to clear it
		return nil, errors.NewAppError(errors.ValidationInvalid, "Register failed", nil).ToServerError(resp.StatusCode)
	}

	// Step 5: Parse API response to extract user data and token
	var response struct {
		User  map[string]interface{} `json:"user"`
		Token string                 `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse registration response")
	}

	// Step 6: Convert response map to User struct
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

	// Step 7: Persist user credentials to local storage
	if err := utils.SetCredentials(&response_user); err != nil {
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
		return &response_user, nil // Don't fail the registration, just return success
	}
	// Initialize the database schema (create tables if they don't exist)
	// Non-fatal operation - registration succeeds even if schema initialization fails
	if err := storage.InitializeSchema(localDB); err != nil {
		fmt.Printf("Warning: Failed to initialize database schema: %v\n", err)
		// Don't fail the registration, just continue
	}

	return &response_user, nil
}
