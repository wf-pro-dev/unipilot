package server

import (
	"context"
	"time"

	//"github.com/gorilla/sessions"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

// LoginHandler handles user authentication requests.
// Validates user credentials against stored password hash and generates JWT tokens
// for authenticated sessions. Uses bcrypt for secure password verification.
//
// HTTP Method: POST
// Content-Type: application/json
//
// Request Body:
//   - username: User's username (string, required)
//   - password: User's password in plain text (string, required)
//
// Response (200 OK):
//   - message: Success message
//   - user: User object (as map) with sensitive fields removed
//   - token: JWT access token (expires in 15 minutes)
//   - refresh_token: JWT refresh token (expires in 30 days)
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON body
//   - 401 Unauthorized: User not found or invalid password
//   - 405 Method Not Allowed: Non-POST request
//   - 500 Internal Server Error: Session key retrieval or token generation failure
//
// Security Features:
//   - Constant-time password comparison using bcrypt
//   - JWT tokens with appropriate expiration times
//   - Structured logging for security audit trails
//
// Side Effects:
//   - Logs authentication attempts (both successful and failed)
//   - Generates new JWT tokens for each login session
func LoginHandler(c *fiber.Ctx) error {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&credentials); err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"tags", []string{"LOGIN", "INVALID_REQUEST_BODY"},
		)
	}

	db := c.Locals("db").(*gorm.DB)

	var userObj user.User
	if err := db.Where("username = ?", credentials.Username).First(&userObj).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusUnauthorized, "No user found",
			"username", credentials.Username,
			"tags", []string{"LOGIN", "USER", "DB"},
		)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(userObj.PasswordHash), []byte(credentials.Password)); err != nil {
		return server.ResponseError(c, err, fiber.StatusUnauthorized, "Invalid Password",
			"user_id", userObj.ID,
			"username", userObj.Username,
			"tags", []string{"LOGIN", "PASSWORD"},
		)
	}

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Invalid session key:",
			"user_id", userObj.ID,
			"username", userObj.Username,
			"tags", []string{"LOGIN", "SESSION_KEY"},
		)
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: userObj,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error creating access token",
			"user_id", userObj.ID,
			"username", userObj.Username,
			"tags", []string{"LOGIN", "ACCESS_TOKEN"},
		)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: userObj,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error creating refresh token",
			"user_id", userObj.ID,
			"username", userObj.Username,
			"tags", []string{"LOGIN", "REFRESH_TOKEN"},
		)
	}

	server.LogInfo(context.Background(), "User logged in", "user_id", userObj.ID,
		"tags", []string{"auth", "auth", "medium", "read"})

	return c.JSON(fiber.Map{
		"message":       "Login successful",
		"user":          userObj.ToMap(),
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
