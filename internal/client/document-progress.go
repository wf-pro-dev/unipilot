package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unipilot/internal/secrets"
	"unipilot/internal/services/fileops/progress"
)

// PollServerProgress polls the server for R2 upload progress
func PollServerProgress(ctx context.Context, uploadID string, tracker *progress.Tracker) {
	api_url := secrets.CONSTANTS["API_URL"]

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	tracker.SetStatus("uploading to cloud storage")

	for {
		select {
		case <-ctx.Done():
			tracker.SetError(ctx.Err())
			return

		case <-ticker.C:
			url := fmt.Sprintf("%s/documents/progress/%s", api_url, uploadID)
			resp, err := http.Get(url)
			if err != nil {
				// Continue polling on error
				continue
			}

			var progressData map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&progressData)
			resp.Body.Close()

			if status, ok := progressData["status"].(string); ok {
				switch status {
				case "completed":
					tracker.Complete()
					return
				case "error":
					tracker.SetError(fmt.Errorf("server upload error"))
					return
				}
			}

			// Update tracker with server progress if available
			if serverProgress, ok := progressData["progress"].(float64); ok {
				// Server progress is 0-100, map it to tracker's total
				tracker.Update(int64(serverProgress * float64(tracker.Total) / 100))
			}

		case <-time.After(60 * time.Second):
			tracker.SetError(fmt.Errorf("server upload timeout"))
			return
		}
	}
}

// CancelUpload cancels an upload on the server
func CancelUpload(uploadID string) error {
	api_url := secrets.CONSTANTS["API_URL"]

	// Cancel local context
	cancel, exists := progress.GetCancelFunc(uploadID)
	if exists {
		cancel()
	}

	// Cancel on server
	url := fmt.Sprintf("%s/documents/upload/%s", api_url, uploadID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	// Set auth header
	if err := SetAuthHeaderRequest(req); err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to cancel upload on server: status %d", resp.StatusCode)
	}

	return nil
}

// GetUploadProgress gets the current progress of an upload from the server
func GetUploadProgress(uploadID string) (map[string]interface{}, error) {
	api_url := secrets.CONSTANTS["API_URL"]

	url := fmt.Sprintf("%s/documents/progress/%s", api_url, uploadID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var progressData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&progressData); err != nil {
		return nil, err
	}

	return progressData, nil
}
