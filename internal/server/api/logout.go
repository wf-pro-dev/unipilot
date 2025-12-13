package server

import (
	"context"

	"unipilot/internal/server"

	"github.com/gofiber/fiber/v2"
)

// LogoutHandler handles user logout requests.
// This is a stateless logout implementation that simply acknowledges the logout request.
// Since the system uses JWT tokens (which are stateless), actual token invalidation
// must be handled client-side by discarding the tokens.
//
// HTTP Method: POST
// Content-Type: Not required (no request body expected)
//
// Request Body: None required
//
// Response (200 OK):
//   - message: Success confirmation message
//
// Security Notes:
//   - This is a client-side logout implementation
//   - JWT tokens remain valid until expiration (15 min for access, 30 days for refresh)
//   - For true server-side logout, consider implementing a token blacklist
//   - Client should discard tokens after receiving this response
//
// Side Effects:
//   - Logs logout event for audit trail
//   - No server-side session state changes (stateless design)
func LogoutHandler(c *fiber.Ctx) error {
	// Step 3: Log logout event for security audit trail and monitoring
	server.LogInfo(context.Background(), "User logged out",
		"tags", []string{"auth", "auth", "low"})

	// Step 2: Send successful logout acknowledgment (no server-side token invalidation needed)
	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}
