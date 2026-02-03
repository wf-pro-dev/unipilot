package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"unipilot/internal/models"
	"unipilot/internal/secrets"
	"unipilot/internal/services/fileops/progress"

	"unipilot/internal/errors"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/datatypes"
)

// DocCreateResp represents the server response
type DocCreateResp struct {
	RemoteID           datatypes.UUID `json:"remote_id"`
	RemoteAssignmentID datatypes.UUID `json:"remote_assignment_id"`
	StorageKey         string         `json:"storage_key"`
}

// GetDocuments retrieves all documents
func GetDocuments() ([]models.Document, error) {

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

	var documents []models.Document
	if err := json.Unmarshal(body, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return documents, nil
}

// GetAssignmentDocuments retrieves documents for a specific assignment
func GetAssignmentDocuments(assignmentID datatypes.UUID) ([]models.Document, error) {
	var response struct {
		Message   string            `json:"message"`
		Documents []models.Document `json:"documents"`
		Error     string            `json:"error,omitempty"`
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

	return response.Documents, nil
}

// sendMultipartFile sends file using multipart/form-data with your authenticated client
func SendDocument(localDocument *models.LocalDocument) (*DocCreateResp, error) {

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
		_, err = io.Copy(fileWriter, fileContent)
		if err != nil {
			return nil, fmt.Errorf("error copying file content: %v", err)
		}

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
	var uploadResp DocCreateResp
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

// SendDocumentWithProgress sends document with progress tracking
func SendDocumentWithProgress(ctx context.Context, localDocument *models.LocalDocument, tracker *progress.Tracker) (*DocCreateResp, error) {
	api_url := secrets.CONSTANTS["API_URL"]
	url := fmt.Sprintf("%s/documents", api_url)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

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

		// Wrap with progress reader
		progressReader := progress.NewReader(fileContent, tracker)

		// Copy file content with progress tracking
		_, err = io.Copy(fileWriter, progressReader)
		if err != nil {
			return nil, fmt.Errorf("error copying file content: %v", err)
		}
	}

	// Add metadata
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

	writer.Close()

	// Create HTTP request with context for cancellation
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := SetAuthHeaderRequest(req); err != nil {
		return nil, err
	}

	// Initialize progress manager (reuse if exists, or create new)
	progressManager := progress.GetManager()

	// Create progress tracker
	connTracker := progressManager.Create(localDocument.ID.String(), localDocument.FileSize)
	connTracker.SetStatus("Connecting to server")
	connTracker.OnProgress(func(t *progress.Tracker) {
		snapshot := t.Snapshot()
		// 40% of the total progress

		if snapshot.Error != nil {
			runtime.EventsEmit(ctx, fmt.Sprintf("upload:error:%s", localDocument.ID.String()), map[string]interface{}{
				"upload_id": localDocument.ID.String(),
				"error":     snapshot.Error.Error(),
			})
		} else {
			snapshot.Percentage = 20 + t.Percentage()*0.4
			runtime.EventsEmit(ctx, fmt.Sprintf("upload:progress:%s", localDocument.ID.String()), snapshot)
		}

	})

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				return &progress.ConnReader{
					Conn:    conn,
					OnWrite: connTracker.Increment,
				}, nil
			},
		},
	}
	go GetProgress(ctx, localDocument.ID.String(), 60)
	resp, err := httpClient.Do(req)
	if err != nil {

		// Check if cancelled
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("upload cancelled")
		}

		progress.GetManager().Remove(localDocument.ID.String())
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var serverError *errors.AppError
		if err := json.Unmarshal(respBody, &serverError); err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse server error")
		}
		return nil, serverError.ToServerError(resp.StatusCode)
	}

	var uploadResp DocCreateResp
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to parse success response")
	}

	return &uploadResp, nil
}

// SendDocument remains for metadata-only uploads (no progress tracking)
// ... keep existing implementation

func DownloadDocument(ctx context.Context, document *models.LocalDocument) (io.Reader, error) {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Post(fmt.Sprintf("%s/documents/%s/download", api_url, document.ID.String()))
	agent.JSON(document)

	if err := SetAuthHeader(agent); err != nil {
		return nil, err
	}

	go GetProgress(ctx, document.ID.String(), 0)
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

func DeleteDocument(documentID datatypes.UUID) error {

	api_url := secrets.CONSTANTS["API_URL"]
	agent := fiber.Delete(fmt.Sprintf("%s/documents/%s", api_url, documentID.String()))

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

func UploadDocumentRAG(document *models.LocalDocument) error {

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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/documents/%s/rag", api_url, document.ID.String()), &buf)
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

func DeleteDocumentRAG(assignmentID, documentID datatypes.UUID) error {

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

func GetAssignmentDocumentIDsRAG(assignmentID datatypes.UUID) ([]datatypes.UUID, error) {

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

	var documentIDs []datatypes.UUID
	if err := json.Unmarshal(body, &documentIDs); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return documentIDs, nil
}
