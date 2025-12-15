package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GenerateChatIDHandler(c *fiber.Ctx) error {
	var chatID uuid.UUID
	chatID = uuid.New()

	return c.JSON(fiber.Map{
		"chat_id": chatID.String(),
	})
}
