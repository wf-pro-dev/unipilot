package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"

	"github.com/gofiber/fiber/v2"
)

func GetUser() (*user.User, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/users/me", api_url))

	if err := setAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var response struct {
		Message string    `json:"message"`
		User    user.User `json:"user"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.User, nil
}

func UpdateUser(column, value string) error {

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/users/me", api_url))
	agent.JSON(updateData)

	if err := setAuthHeader(agent); err != nil {
		return err
	}

	statusCode, _, errs := agent.Bytes()
	if len(errs) > 0 {
		return errs[0]
	}

	if statusCode != 200 {
		return fmt.Errorf("server returned status %d", statusCode)
	}

	return nil
}

type ProfilePictureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func UpdateProfilePicture(path string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	var url string = fmt.Sprintf("%s/users/me/profile-picture", api_url)

	// Create a buffer to store the multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	defer writer.Close()

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", path)
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}

	// Open the file
	fileContent, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fileContent.Close()

	// Copy file content to form
	bytesWritten, err := io.Copy(fileWriter, fileContent)
	if err != nil {
		return fmt.Errorf("error copying file content: %v", err)
	}

	log.Printf("bytes written: %d\n", bytesWritten)

	// Close the writer to finalize the multipart message
	writer.Close()

	// Create HTTP request using your server URL
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers - the Content-Type is crucial for multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set auth header with token refresh
	if err := setAuthHeaderRequest(req); err != nil {
		return err
	}

	// Send request using default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("response: %v", resp.StatusCode)

	// Handle different HTTP status codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// Success - parse response
		var uploadResp ProfilePictureResponse
		if err := json.Unmarshal(respBody, &uploadResp); err != nil {
			return fmt.Errorf("error parsing success response: %v", err)
		}
		return nil

	case http.StatusUnauthorized:
		return fmt.Errorf("authentication required. Please login first")

	case http.StatusForbidden:
		return fmt.Errorf("access forbidden. You don't have permission to upload files")

	case http.StatusRequestEntityTooLarge:
		return fmt.Errorf("file too large for server")

	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("file type not supported")

	default:
		return fmt.Errorf("server error: %s - %s", resp.Status, string(respBody))
	}
}
