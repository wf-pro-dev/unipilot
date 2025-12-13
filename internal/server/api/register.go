package server

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

// RegisterHandler handles user registration requests.
// Creates a new user account with hashed password, generates JWT tokens,
// and caches user data in Redis for performance optimization.
//
// HTTP Method: POST
// Content-Type: application/json
//
// Request Body:
//   - username: User's chosen username (string, required)
//   - email: User's email address (string, required)
//   - password: User's password in plain text (string, required)
//   - university: User's university affiliation (string, required)
//   - language: User's preferred language (string, required)
//
// Response (200 OK):
//   - message: Success message
//   - user: User object (as map) with sensitive fields removed
//   - token: JWT access token (expires in 15 minutes)
//   - refresh_token: JWT refresh token (expires in 30 days)
//
// Error Responses:
//   - 400 Bad Request: Invalid JSON body or missing required fields
//   - 405 Method Not Allowed: Non-POST request
//   - 409 Conflict: Username already exists
//   - 500 Internal Server Error: Password hashing, database, or token generation failure
//
// Side Effects:
//   - Creates new user record in database
//   - Caches user data in Redis (non-blocking, logs warning on failure)
//   - Generates cryptographically secure password hash using bcrypt
func RegisterHandler(c *fiber.Ctx) error {
	// Step 2: Define input structure for JSON unmarshaling with required user fields
	var registrationData struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		University string `json:"university"`
		Language   string `json:"language"`
	}

	// Parse JSON request body into registration data structure
	err := c.BodyParser(&registrationData)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusBadRequest, "Invalid request body",
			"tags", []string{"REGISTER"},
		)
	}

	// Extract database connection from middleware context
	db := c.Locals("db").(*gorm.DB)

	// Step 3: Validate all required fields are present (business rule enforcement)
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.University == "" || registrationData.Language == "" {
		return server.ResponseError(c, errors.New("missing required fields"), fiber.StatusBadRequest, "Username, email, password, university, and language are required",
			"tags", []string{"REGISTER", "MISSING_REQUIRED_FIELDS"},
		)
	}

	// Step 4: Check username uniqueness constraint (prevents duplicate accounts)
	if err := db.Where("username = ?", registrationData.Username).First(&user.User{}).Error; err == nil {
		return server.ResponseError(c, errors.New("username already taken"), fiber.StatusConflict, "Username already taken",
			"tags", []string{"REGISTER", "DB"},
		)
	}

	// Step 5: Hash password using bcrypt with default cost (currently 10 rounds)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error generating hashed password",
			"tags", []string{"REGISTER", "PASSWORD"},
		)
	}

	// Step 6: Construct user object with validated data and hashed password
	userObj := user.User{
		Username:     registrationData.Username,
		Email:        registrationData.Email,
		PasswordHash: string(hashedPassword),
		University:   registrationData.University,
		Language:     registrationData.Language,
	}

	// Persist new user to database
	if err := db.Create(&userObj).Error; err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error creating user",
			"tags", []string{"REGISTER", "DB"},
		)
	}

	// Convert user struct to map for safe JSON response (removes sensitive fields)
	userMap := userObj.ToMap()

	// Step 7: Generate JWT tokens for immediate authentication after registration
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error getting session key",
			"tags", []string{"REGISTER", "SESSION_KEY"},
		)
	}

	// Create short-lived access token (15 minutes) for API access
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
			"tags", []string{"REGISTER", "ACCESS_TOKEN"},
		)
	}

	// Create long-lived refresh token (30 days) for token renewal
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
			"tags", []string{"REGISTER", "REFRESH_TOKEN"},
		)
	}

	// Step 8: Cache user data in Redis for performance optimization (non-blocking)
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		return server.ResponseError(c, err, fiber.StatusInternalServerError, "Error marshalling user to json",
			"user_id", userObj.ID,
			"tags", []string{"REGISTER", "MARSHALLING"},
		)
	}

	// Store in Redis hash with user ID as key (failure is non-critical, only logged)
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(userObj.ID)), userJSON).Err(); err != nil {
		server.LogWarn(context.Background(), "Failed to cache user in Redis", err, "user_id", userObj.ID,
			"tags", []string{"cache", "cache", "medium"},
			"cache_status", "error",
			"error_type", "cache",
		)
	}

	// Step 10: Log successful registration for audit trail and monitoring
	server.LogInfo(context.Background(), "User registered", "user_id", userObj.ID,
		"tags", []string{"user", "db", "high", "create"})

	// Step 9: Send successful registration response with user data and tokens
	return c.JSON(fiber.Map{
		"message":       "Register successful",
		"user":          userMap,
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}
