package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"unipilot/internal/secrets"
	"unipilot/internal/services/fileops/progress"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PollServerProgress polls the server for R2 upload progress
func PollServerProgress(ctx context.Context, uploadID string, tracker *progress.Tracker) {
	api_url := secrets.CONSTANTS["API_URL"]

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	tracker.SetStatus("uploading to cloud storage")
	log.Printf("Starting Upload ID: %s", uploadID)
	for {
		select {
		case <-ctx.Done():
			tracker.SetError(ctx.Err())
			return

		case <-ticker.C:
			url := fmt.Sprintf("%s/documents/progress/%s", api_url, uploadID)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				tracker.SetError(err)
			}
			if err := SetAuthHeaderRequest(req); err != nil {
				tracker.SetError(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// Continue polling on error
				tracker.SetError(err)
			}

			var progressData progress.TrackerSnapshot
			if err := json.NewDecoder(resp.Body).Decode(&progressData); err != nil {
				tracker.SetError(err)
			}
			resp.Body.Close()

			switch progressData.Status {
			case "completed":
				runtime.EventsEmit(ctx, "upload:complete", map[string]interface{}{
					"upload_id": uploadID,
				})
				return
			case "error":
				runtime.EventsEmit(ctx, "upload:error", map[string]interface{}{
					"upload_id": uploadID,
					"error":     "server upload error",
				})
				tracker.SetError(fmt.Errorf("server upload error"))
				return
			default:
				progressData.Percentage = 20 + progressData.Percentage*0.8 // 80% of the total progress				runtime.EventsEmit(ctx, "upload:progress", progressData)

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
