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
	"unipilot/internal/secrets"
)

func SendUserUpdate(column, value string) error {

	new_client, err := NewAuthClient()
	if err != nil {

		return err
	}

	updateData := map[string]interface{}{
		"value":  value,
		"column": column,
	}

	jsonData, _ := json.Marshal(updateData)

	api_url := secrets.CONSTANTS["API_URL"]
	resp, err := new_client.Post(
		fmt.Sprintf("%s/user/update", api_url),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type ProfilePictureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendProfilePicture(path string) error {

	api_url := secrets.CONSTANTS["API_URL"]
	var url string = fmt.Sprintf("%s/user/profile-picture", api_url)

	// Create a new client with cookies
	client, err := NewAuthClient()
	if err != nil {
		return err
	}

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

	// Send request using your authenticated client with cookies
	resp, err := client.Do(req)
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
		var uploadResp UploadResponse
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
