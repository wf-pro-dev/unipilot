package server

import (
	"fmt"
	"time"

	//"github.com/gorilla/sessions"
	Errors "errors"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unipilot/internal/models"
	"unipilot/internal/secrets"
	"unipilot/internal/server"

	"unipilot/internal/errors"
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
	c.Locals("message", "User logging in")
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&credentials); err != nil {
		return errors.WrapServer(err, errors.ReqBodyInvalid, "Invalid request body", fiber.StatusBadRequest)
	}

	db, ok := c.Locals("db").(*gorm.DB)
	if !ok {
		return errors.WrapServer(fmt.Errorf("database connection not found"), errors.DBConnectionFailed, "Database connection not found", fiber.StatusInternalServerError)
	}

	var userObj models.User
	if err := db.Where("username = ?", credentials.Username).First(&userObj).Error; err != nil {
		if Errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapServer(err, errors.DBRecordNotFound, "No user found", fiber.StatusUnauthorized)
		}
		return errors.WrapServer(err, errors.DBQueryFailed, "No user found", fiber.StatusUnauthorized)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(userObj.PasswordHash), []byte(credentials.Password)); err != nil {
		return errors.WrapServer(err, errors.AuthUnauthorized, "Invalid Password", fiber.StatusUnauthorized)
	}

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		return errors.WrapServer(err, errors.ConfigEnvVarNotFound, "Invalid session key:", fiber.StatusInternalServerError)
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: userObj,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		return errors.WrapServer(err, errors.AuthTokenGeneration, "Error creating access token", fiber.StatusInternalServerError)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: userObj,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		return errors.WrapServer(err, errors.AuthTokenGeneration, "Error creating refresh token", fiber.StatusInternalServerError)
	}

	c.Locals("message", "User logged in")

	return c.JSON(fiber.Map{
		"message":       "Login successful",
		"user":          userObj.ToMap(),
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
