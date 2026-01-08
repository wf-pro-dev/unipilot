package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
	"unipilot/internal/errors"
	"unipilot/internal/services/fileops/progress"

	"github.com/gofiber/fiber/v2"
)

// GetUploadProgressHandler returns the current progress of an upload
// GetUploadProgressHandler returns the current progress of an upload
// GetUploadProgressHandler returns the current progress of an upload
func GetUploadProgressHandler(c *fiber.Ctx) error {
	c.Locals("message", "Upload progress fetched successfully")
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return errors.WrapServer(
			fmt.Errorf("upload_id parameter required"),
			errors.ReqParamMissing,
			"Upload ID required",
			fiber.StatusBadRequest,
		)
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("X-Accel-Buffering", "no")

	// Get progress from Redis
	redisPubSub, err := CacheService.GetProgressChannel(c.UserContext(), uploadID)
	if err != nil {
		return errors.WrapServer(
			err,
			errors.CacheOperationFailed,
			"Failed to get upload progress",
			fiber.StatusInternalServerError,
		)
	}

	initialProgress := &progress.TrackerSnapshot{
		ID:         uploadID,
		Status:     "stopped",
		Current:    0,
		Total:      0, // Unknown at this point
		Error:      nil,
		Percentage: 0,
	}
	CacheService.PublishProgress(c.UserContext(), uploadID, initialProgress)

	// Use c.Context() directly for Done() channel
	ctx := c.Context()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			redisPubSub.Close()
			w.Flush()
		}()

		progressChannel := redisPubSub.Channel()

		// Send initial connection message
		fmt.Fprintf(w, "event: connected\ndata: Connected to progress stream\n\n")
		w.Flush()

		timeout := time.NewTimer(2 * time.Minute)
		defer timeout.Stop()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Client disconnected
				return
			case <-ticker.C:
				// Send keep-alive comment
				fmt.Fprintf(w, ": keepalive\n\n")
				w.Flush()
			case <-timeout.C:
				// Send timeout message after 5 minutes
				timeoutProgress := &progress.TrackerSnapshot{
					ID:         uploadID,
					Status:     "error",
					Error:      fmt.Errorf("Upload did not complete within timeout period"),
					Current:    0,
					Total:      0,
					Percentage: 0,
				}
				data, _ := json.Marshal(timeoutProgress)
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.Flush()
				return
			case msg, ok := <-progressChannel:

				if !ok || msg == nil {
					continue
				}
				timeout.Reset(2 * time.Minute)
				// Send progress data
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				w.Flush()
			}
		}
	})

	return nil
}
