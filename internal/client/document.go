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
	"unipilot/internal/models/document"
	"unipilot/internal/secrets"
)

// UploadResponse represents the server response
type UploadResponse struct {
	Success  bool              `json:"success"`
	Document document.Document `json:"document"`
}

// sendMultipartFile sends file using multipart/form-data with your authenticated client
func SendDocument(document *document.LocalDocument) (string, error) {

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return "", fmt.Errorf("failed to get api url: %w", err)
	}

	var url string = fmt.Sprintf("%s/document", api_url)

	// Create a new client with cookies
	client, err := NewClientWithCookies()
	if err != nil {
		return "", err
	}

	// Create a buffer to store the multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	defer writer.Close()

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", document.FileName)
	if err != nil {
		return "", fmt.Errorf("error creating form file: %v", err)
	}

	if document.HasLocalFile {

		// Open the file
		fileContent, err := os.Open(document.FilePath)
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer fileContent.Close()

		// Copy file content to form
		bytesWritten, err := io.Copy(fileWriter, fileContent)
		if err != nil {
			return "", fmt.Errorf("error copying file content: %v", err)
		}

		log.Printf("bytes written: %d\n", bytesWritten)
	}

	// Add metadata part
	metadataWriter, err := writer.CreateFormField("metadata")
	if err != nil {
		return "", fmt.Errorf("error creating form field: %v", err)
	}

	metadataJSON, err := json.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("error marshalling metadata: %v", err)
	}

	_, err = metadataWriter.Write(metadataJSON)
	if err != nil {
		return "", fmt.Errorf("error writing metadata: %v", err)
	}

	// Close the writer to finalize the multipart message
	writer.Close()

	// Create HTTP request using your server URL
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers - the Content-Type is crucial for multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request using your authenticated client with cookies
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	log.Printf("response: %v", resp.StatusCode)

	// Handle different HTTP status codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		// Success - parse response
		var uploadResp UploadResponse
		if err := json.Unmarshal(respBody, &uploadResp); err != nil {
			return "", fmt.Errorf("error parsing success response: %v", err)
		}
		return uploadResp.Document.StorageKey, nil

	case http.StatusUnauthorized:
		return "", fmt.Errorf("authentication required. Please login first")

	case http.StatusForbidden:
		return "", fmt.Errorf("access forbidden. You don't have permission to upload files")

	case http.StatusRequestEntityTooLarge:
		return "", fmt.Errorf("file too large for server")

	case http.StatusUnsupportedMediaType:
		return "", fmt.Errorf("file type not supported")

	default:
		return "", fmt.Errorf("server error: %s - %s", resp.Status, string(respBody))
	}
}
func DownloadDocument(document *document.LocalDocument) (io.Reader, error) {

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return nil, fmt.Errorf("failed to get api url: %w", err)
	}

	var url string = fmt.Sprintf("%s/document/download", api_url)

	jsonData, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}

	client, err := NewClientWithCookies()
	if err != nil {
		return nil, err
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the error message from the body
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/octet-stream" {
		return nil, fmt.Errorf("unexpected content type: %s", contentType)
	}

	// Read the entire content into a buffer and return it
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &buf, nil
}

func DeleteDocument(documentID uint) error {

	api_url, err := secrets.GetEnvVar("API_URL")
	if err != nil {
		return fmt.Errorf("failed to get api url: %w", err)
	}

	var url string = fmt.Sprintf("%s/document/delete", api_url)

	client, err := NewClientWithCookies()
	if err != nil {
		return err
	}

	resp, err := client.Post(fmt.Sprintf("%s?document_id=%d", url, documentID), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}
