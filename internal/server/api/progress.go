package server

import (
	"fmt"
	"log"
	"time"
	"unipilot/internal/errors"
	"unipilot/internal/services/fileops/progress"

	"github.com/gofiber/fiber/v2"
)

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

	var progressData *progress.TrackerSnapshot
	var err error

	// Get progress from Redis
	progressData, err = CacheService.GetProgress(c.Context(), uploadID)
	if err != nil {
		if errors.HasCode(err, errors.CacheMiss) {
			progressData = &progress.TrackerSnapshot{
				ID:         uploadID,
				Total:      0,
				Current:    0,
				Status:     "Uploading to server",
				StartTime:  time.Now(),
				Error:      nil,
				Percentage: 0,
			}
		} else {
			return errors.WrapServer(
				err,
				errors.CacheOperationFailed,
				"Failed to get upload progress",
				fiber.StatusInternalServerError,
			)
		}
	}

	log.Printf("Progress data: %d, %f, %s", progressData.Current, progressData.Percentage, progressData.Status)

	return c.JSON(progressData)
}
