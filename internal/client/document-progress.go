package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unipilot/internal/errors"
	"unipilot/internal/secrets"
	"unipilot/internal/services/fileops/progress"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PollServerProgress polls the server for R2 upload progress
func PollServerProgress(ctx context.Context, progressID string, tracker *progress.Tracker) {
	api_url := secrets.CONSTANTS["API_URL"]

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			tracker.SetError(ctx.Err())
			return

		case <-ticker.C:
			url := fmt.Sprintf("%s/progress/%s", api_url, progressID)
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
				runtime.EventsEmit(ctx, fmt.Sprintf("upload:complete:%s", progressID), map[string]interface{}{
					"upload_id": progressID,
				})
				return
			case "error":
				runtime.EventsEmit(ctx, fmt.Sprintf("upload:error:%s", progressID), map[string]interface{}{
					"upload_id": progressID,
					"error":     "server upload error",
				})
				tracker.SetError(fmt.Errorf("server upload error"))
				return
			default:
				progressData.Percentage = 60 + progressData.Percentage*0.4 // 40% of the total progress
				runtime.EventsEmit(ctx, fmt.Sprintf("upload:progress:%s", progressID), progressData)

			}

		case <-time.After(60 * time.Second):
			tracker.SetError(fmt.Errorf("server upload timeout"))
			return
		}
	}
}

// CancelUpload cancels an upload on the server
// GetProgress listens for upload progress via SSE
func GetProgress(ctx context.Context, progressID string, currentPercentage float64) error {
	api_url := secrets.CONSTANTS["API_URL"]
	url := fmt.Sprintf("%s/progress/%s", api_url, progressID)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, errors.ClientRequestFailed, "Failed to create request")
	}

	if err := SetAuthHeaderRequest(req); err != nil {
		return err
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Create client WITHOUT custom transport for SSE
	// SSE requires persistent connection
	httpClient := &http.Client{
		Timeout: 0, // No timeout for SSE streams
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.NetworkConnectionFailed, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return errors.NewAppError(
			errors.ClientResponseInvalid,
			fmt.Sprintf("Received non-200 status code: %d - %s", resp.StatusCode, string(body)),
			nil,
		)
	}

	// Create SSE reader
	reader := bufio.NewReader(resp.Body)
	var eventType string
	var dataBuffer strings.Builder

	for {
		// Check context
		select {
		case <-ctx.Done():
			log.Println("Context done", ctx.Err())
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("EOF")
				return nil // Normal stream end
			}
			log.Println("Error reading from SSE stream:", err)
			return errors.Wrap(err, errors.FSStreamFailed, "Error reading from SSE stream")
		}

		line = strings.TrimSpace(line)

		if line == "" {
			// Empty line indicates end of event
			if dataBuffer.Len() > 0 {
				data := dataBuffer.String()

				// Handle based on event type
				switch eventType {
				case "connected":
					log.Println("Connected to SSE stream")
				case "error":
					log.Println("Server error:", data)
					runtime.EventsEmit(ctx, fmt.Sprintf("upload:error:%s", progressID), map[string]interface{}{
						"upload_id": progressID,
						"error":     data,
					})
					return fmt.Errorf("server error: %s", data)
				default:

					// Parse progress data
					var progressData progress.TrackerSnapshot
					if err := json.Unmarshal([]byte(data), &progressData); err != nil {
						log.Printf("Failed to unmarshal progress data: %v %s", err, data)
						return errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal progress data")
					}

					// Check for completion
					if progressData.Status == "completed" {
						runtime.EventsEmit(ctx, fmt.Sprintf("upload:complete:%s", progressID), map[string]interface{}{
							"upload_id": progressID,
						})
						return nil
					}
					if progressData.Status == "error" {
						log.Println("Server upload error:", progressData.Error)
						return fmt.Errorf("server upload error: %s", progressData.Error)
					}

					if progressData.Status != "stopped" {
						// Adjust percentage if needed
						progressData.Percentage = currentPercentage + progressData.Percentage*(100-currentPercentage)/100
						// Emit progress
						runtime.EventsEmit(ctx, fmt.Sprintf("upload:progress:%s", progressID), progressData)
					}

				}

				// Reset for next event
				dataBuffer.Reset()
				eventType = ""
			}
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if dataBuffer.Len() > 0 {
				dataBuffer.WriteString("\n")
			}
			dataBuffer.WriteString(data)
		} else if strings.HasPrefix(line, ":") {
			// Comment line, ignore
			log.Println("Comment line:", line)
			continue
		}
	}
}
