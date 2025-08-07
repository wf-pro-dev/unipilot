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
)

// Login handles only authentication and saving the session cookie to a file.
func (a *Auth) Register(username, email, password, university, language string) error {

	httpClient, err := client.NewClientWithCookies() // Changed from NewClient()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	// Set the client to the auth struct
	a.Client = httpClient

	loginData := map[string]string{"username": username, "password": password, "email": email, "university": university, "language": language}
	jsonData, _ := json.Marshal(loginData)

	resp, err := httpClient.Post("https://newsroom.dedyn.io/acc-homework/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Register response status code: ", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := client.SaveCookies(httpClient); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	// Parse the response to get user ID
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	fmt.Printf("Response: %v\n", response)

	// Extract user ID from response
	userIDStr, ok := response["id"].(string)
	if !ok {
		return fmt.Errorf("invalid id in response")
	}
	userIDint, err := strconv.Atoi(userIDStr)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	fmt.Printf("User ID: %v\n", userIDint)

	// Store credentials first
	if err := storage.StoreCredentials(uint(userIDint), username); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	// Initialize the SSE connection early to ensure it's never nil
	a.SSE = sse.NewSSE()

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

	return nil
}
