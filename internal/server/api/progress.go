package server

import (
	"fmt"
	"unipilot/internal/errors"

	"github.com/gofiber/fiber/v2"
)

// GetUploadProgressHandler returns the current progress of an upload
func GetUploadProgressHandler(c *fiber.Ctx) error {
	uploadID := c.Params("upload_id")
	if uploadID == "" {
		return errors.WrapServer(
			fmt.Errorf("upload_id parameter required"),
			errors.ReqParamMissing,
			"Upload ID required",
			fiber.StatusBadRequest,
		)
	}

	// Get progress from Redis
	progressData, err := CacheService.GetProgress(c.Context(), uploadID)
	if err != nil {
		if errors.HasCode(err, errors.CacheOperationFailed) {
			return fiber.ErrNotFound
		}
		return errors.WrapServer(
			err,
			errors.CacheOperationFailed,
			"Failed to get upload progress",
			fiber.StatusInternalServerError,
		)
	}

	return c.JSON(progressData)
}
