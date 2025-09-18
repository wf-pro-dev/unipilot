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
	"unipilot/internal/services/fileops"
)

// UploadResponse represents the server response
type UploadResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	FileURL    string `json:"file_url,omitempty"`
	FileName   string `json:"file_name,omitempty"`
	Size       int64  `json:"size,omitempty"`
	DocumentID string `json:"document_id,omitempty"`
}

// sendMultipartFile sends file using multipart/form-data with your authenticated client
func SendDocument(file *fileops.FileUploadRequest, localID uint) error {

	var url string = "https://newsroom.dedyn.io/acc-homework/document/metadata"

	// Create a new client with cookies
	client, err := NewClientWithCookies()
	if err != nil {
		return err
	}

	// Create a buffer to store the multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", file.FileName)
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}

	// Open the file
	fileContent, err := os.Open(file.FilePath)
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

	// Add additional metadata fields
	writer.WriteField("assignment_id", fmt.Sprintf("%d", file.AssignmentID))
	writer.WriteField("local_id", fmt.Sprintf("%d", localID))
	writer.WriteField("user_id", fmt.Sprintf("%d", file.UserID))
	writer.WriteField("type", string(file.Type))
	writer.WriteField("file_name", file.FileName)
	writer.WriteField("file_type", fileops.GetMimeType(file.FileName))
	writer.WriteField("file_size", fmt.Sprintf("%d", file.FileSize))

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
