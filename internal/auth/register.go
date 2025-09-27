package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/models/user"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
	"unipilot/internal/storage"
)

// Login handles only authentication and saving the session cookie to a file.
func (a *Auth) Register(username, email, password, university, language string) (*user.User, error) {

	httpClient, err := client.NewClientWithCookies() // Changed from NewClient()
	if err != nil {
		return nil, fmt.Errorf("could not create http client: %w", err)
	}

	// Set the client to the auth struct
	a.Client = httpClient

	loginData := map[string]string{"username": username, "password": password, "email": email, "university": university, "language": language}
	jsonData, _ := json.Marshal(loginData)

	resp, err := httpClient.Post("https://newsroom.dedyn.io/acc-homework/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Register response status code: ", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("register failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := client.SaveCookies(httpClient); err != nil {
		return nil, fmt.Errorf("failed to save cookies: %w", err)
	}

	// Parse the response to get user ID
	var response struct {
		User map[string]interface{} `json:"user"`
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

	// Store the user in credentials
	if err := utils.SetCredentials(&response_user); err != nil {
		return nil, fmt.Errorf("failed to set credentials: %w", err)
	}

	// Initialize the SSE connection early to ensure it's never nil
	a.SSE = sse.NewSSE()

	// Now try to get the local database and migrate data
	// But handle the case where it might fail gracefully
	localDB, err := utils.GetUserDB()
	if err != nil {
		// If we can't get the local database, just log it and continue
		// This might happen if the database directory doesn't exist yet
		fmt.Printf("Warning: Could not get local database: %v\n", err)
		fmt.Printf("Login successful, but database operations failed\n")
		return &response_user, nil // Don't fail the login, just return success
	}
	// Initialize the database schema
	if err := storage.InitializeSchema(localDB); err != nil {
		fmt.Printf("Warning: Failed to initialize database schema: %v\n", err)
		// Don't fail the login, just continue
	}

	return &response_user, nil
}
