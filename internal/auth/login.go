package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"unipilot/internal/client"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
	"unipilot/internal/sync"
)

// Login handles only authentication and saving the session cookie to a file.
func (a *Auth) Login(username, password string) error {

	httpClient, err := client.NewClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	// Set the client to the auth struct
	a.Client = httpClient

	loginData := map[string]string{"username": username, "password": password}
	jsonData, _ := json.Marshal(loginData)

	resp, err := httpClient.Post("https://newsroom.dedyn.io/acc-homework/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := client.SaveCookies(httpClient); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	// Parse the response to get user ID
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract user ID from response
	userIDStr, ok := response["user_id"].(string)
	if !ok {
		return fmt.Errorf("invalid user_id in response")
	}
	userIDint, err := strconv.Atoi(userIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	// Store credentials first
	if err := storage.StoreCredentials(uint(userIDint), username); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	// Now try to get the local database and migrate data
	// But handle the case where it might fail gracefully
	localDB, _, err := storage.GetLocalDB()
	if err != nil {
		// If we can't get the local database, just log it and continue
		// This might happen if the database directory doesn't exist yet
		fmt.Printf("Warning: Could not get local database: %v\n", err)
		fmt.Printf("Login successful, but database operations failed\n")
		return nil // Don't fail the login, just return success
	}

	a.LocalDB = localDB

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

	// Initialize the SSE connection
	a.SSE = sse.NewSSE()

	return nil
}
