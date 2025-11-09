package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
	"unipilot/internal/sync"
)

// Login handles only authentication and saving the session cookie to a file.
func (a *Auth) Login(username, password string) (*user.User, error) {

	// Cet a authenticated client
	httpClient := http.Client{
		Jar: &client.CookieJar{},
	}

	loginData := map[string]string{"username": username, "password": password}
	jsonData, _ := json.Marshal(loginData)

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	resp, err := httpClient.Post(fmt.Sprintf("%s/login", api_url), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status %d", resp.StatusCode)
	}

	// Parse the response to get user ID
	var response struct {
		User         map[string]interface{} `json:"user"`
		Token        string                 `json:"token"`
		RefreshToken string                 `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	response_user := user.User{
		Username:   response.User["username"].(string),
		Email:      response.User["email"].(string),
		Avatar:     response.User["avatar"].(string),
		University: response.User["university"].(string),
		Semester:   response.User["semester"].(string),
		Year:       response.User["year"].(string),
		Language:   response.User["language"].(string),
	}

	response_user.CreatedAt, _ = time.Parse(time.RFC3339, response.User["created_at"].(string))
	response_user.UpdatedAt, _ = time.Parse(time.RFC3339, response.User["updated_at"].(string))

	response_user.ID = uint(response.User["id"].(float64))

	//Store the user in credentials
	if err := utils.SetCredentials(&response_user); err != nil {
		return nil, fmt.Errorf("failed to set credentials: %w", err)
	}

	/* DEPRECATED
	if err := client.SaveCookies(&httpClient); err != nil {
		return nil, fmt.Errorf("failed to save cookies: %w", err)
	}*/

	if err := client.SaveToken(response.Token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	if err := client.SaveRefreshToken(response.RefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	httpUserClient, err := client.NewAuthClient()
	if err != nil {
		return nil, fmt.Errorf("could not create authenticated http client: %w", err)
	}

	a.Client = httpUserClient

	// Initialize the SSE connection early to ensure it's never nil
	a.SSE = sse.NewSSE()

	// Now try to get the local database and migrate data
	// But handle the case where it might fail gracefully

	if err := PostLogin(); err != nil {
		return &response_user, err
	}

	return &response_user, nil
}

func PostLogin() error {
	localDB, err := utils.GetUserDB()
	if err != nil {
		// If we can't get the local database, just log it and continue
		// This might happen if the database directory doesn't exist yet
		fmt.Printf("Warning: Could not get local database: %v\n", err)
		fmt.Printf("Login successful, but database operations failed\n")
		// Don't fail the login, just return success
	}
	// Initialize the database schema
	if err := storage.InitializeSchema(localDB); err != nil {
		fmt.Printf("Warning: Failed to initialize database schema: %v\n", err)
		// Don't fail the login, just continue
	}

	// Try to migrate courses, but don't fail if it doesn't work
	// Note: Sync functions are temporarily disabled
	if err := sync.MigrateCourses(localDB); err != nil {
		fmt.Printf("Warning: Failed to migrate courses: %v\n", err)
		// Don't rollback, continue with the transaction
	}

	//Try to migrate assignments, but don't fail if it doesn't work
	//Note: Sync functions are temporarily disabled
	fmt.Printf("Attempting to migrate assignments...\n")
	if err := sync.MigrateAssignments(localDB); err != nil {
		fmt.Printf("Warning: Failed to migrate assignments: %v\n", err)
		// Don't rollback, continue with the transaction
	}
	return err
}
