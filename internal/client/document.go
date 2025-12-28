package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"unipilot/internal/models/document"
	"unipilot/internal/secrets"

	"unipilot/internal/errors"

	"log"

	"github.com/gofiber/fiber/v2"
)

// UploadResponse represents the server response
type UploadResponse struct {
	RemoteID           uint   `json:"remote_id"`
	RemoteAssignmentID uint   `json:"remote_assignment_id"`
	StorageKey         string `json:"storage_key"`
}

// GetDocuments retrieves all documents
func GetDocuments() ([]document.Document, error) {
	var response struct {
		Message   string              `json:"message"`
		Documents []document.Document `json:"documents"`
		Error     string              `json:"error,omitempty"`
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/documents", api_url))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf(response.Error)
	}

	return response.Documents, nil
}

// GetAssignmentDocuments retrieves documents for a specific assignment
func GetAssignmentDocuments(assignmentID uint) ([]document.Document, error) {
	var response struct {
		Message   string              `json:"message"`
		Documents []document.Document `json:"documents"`
		Error     string              `json:"error,omitempty"`
	}

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/documents/assignments/%d", api_url, assignmentID))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf(response.Error)
	}

	return response.Documents, nil
}

// sendMultipartFile sends file using multipart/form-data with your authenticated client
func SendDocument(localDocument *document.LocalDocument) (*UploadResponse, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	var url string = fmt.Sprintf("%s/documents", api_url)

	// Create a buffer to store the multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	defer writer.Close()

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", localDocument.FileName)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %v", err)
	}

	if localDocument.HasLocalFile {

		// Open the file
		fileContent, err := os.Open(localDocument.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer fileContent.Close()

		// Copy file content to form
		bytesWritten, err := io.Copy(fileWriter, fileContent)
		if err != nil {
			return nil, fmt.Errorf("error copying file content: %v", err)
		}
		log.Printf("bytes written: %d", bytesWritten)
	}

	// Add metadata part
	metadataWriter, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, fmt.Errorf("error creating form field: %v", err)
	}

	metadataJSON, err := json.Marshal(localDocument)
	if err != nil {
		return nil, fmt.Errorf("error marshalling metadata: %v", err)
	}

	_, err = metadataWriter.Write(metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("error writing metadata: %v", err)
	}

	// Close the writer to finalize the multipart message
	writer.Close()

	// Create HTTP request using your server URL
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers - the Content-Type is crucial for multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set auth header with token refresh
	if err := SetAuthHeaderRequest(req); err != nil {
		return nil, err
	}

	// Send request using default client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Success - parse response
	var uploadResp UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse success response")
	}

	if resp.StatusCode != http.StatusOK {
		var serverError *errors.AppError
		if err := json.Unmarshal(respBody, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError.ToServerError(resp.StatusCode)
	}

	return &uploadResp, nil

}
func DownloadDocument(document *document.LocalDocument) (io.Reader, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/documents/%d/download", api_url, document.ID))
	agent.JSON(document)

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	// Read the entire content into a buffer and return it
	var buf bytes.Buffer
	buf.Write(body)

	return &buf, nil
}

func DeleteDocument(documentID uint) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/documents/%d", api_url, documentID))

	if err := SetAuthHeader(agent); err != nil {
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

func UploadDocumentRAG(document *document.LocalDocument) error {

	// Create a buffer to store the multipart data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	defer writer.Close()

	// Create form file field
	fileWriter, err := writer.CreateFormFile("file", document.FileName)
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}

	if document.HasLocalFile {

		// Open the file
		fileContent, err := os.Open(document.FilePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fileContent.Close()

		// Copy file content to form
		_, err = io.Copy(fileWriter, fileContent)
		if err != nil {
			return fmt.Errorf("error copying file content: %v", err)
		}

	}

	// Add metadata part
	metadataWriter, err := writer.CreateFormField("metadata")
	if err != nil {
		return fmt.Errorf("error creating form field: %v", err)
	}

	metadataJSON, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %v", err)
	}

	_, err = metadataWriter.Write(metadataJSON)
	if err != nil {
		return fmt.Errorf("error writing metadata: %v", err)
	}

	// Close the writer to finalize the multipart message
	writer.Close()

	// Create HTTP request using your server URL
	api_url := secrets.CONSTANTS["API_URL"]
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/documents/%d/rag", api_url, document.ID), &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers - the Content-Type is crucial for multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set auth header with token refresh
	if err := SetAuthHeaderRequest(req); err != nil {
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

	if resp.StatusCode != http.StatusOK {
		var serverError *errors.AppError

		if err := json.Unmarshal(respBody, &serverError); err != nil {
			return errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		log.Printf("error body: %s", serverError)
		return serverError.ToServerError(resp.StatusCode)
	}

	return nil
}

func DeleteDocumentRAG(assignmentID, documentID uint) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/documents/%d/%d/rag", api_url, documentID, assignmentID))

	if err := SetAuthHeader(agent); err != nil {
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

func GetAssignmentDocumentIDsRAG(assignmentID uint, documentIDs []uint) ([]uint, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Get(fmt.Sprintf("%s/documents/assignments/%d/rag", api_url, assignmentID))

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("server returned status %d: %s", statusCode, string(body))
	}

	var res struct {
		DocumentIDs []uint `json:"document_ids"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return res.DocumentIDs, nil
}
