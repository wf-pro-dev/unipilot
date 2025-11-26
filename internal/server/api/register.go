package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce POST-only endpoint for security (registration should never be GET)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Define input structure for JSON unmarshaling with required user fields
	var registrationData struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		University string `json:"university"`
		Language   string `json:"language"`
	}

	// Parse JSON request body into registration data structure
	err := json.NewDecoder(r.Body).Decode(&registrationData)
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"REGISTER"},
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract database connection from middleware context
	db := r.Context().Value("db").(*gorm.DB)

	// Step 3: Validate all required fields are present (business rule enforcement)
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.University == "" || registrationData.Language == "" {
		server.ResponseError(r.Context(),
			w, err, http.StatusBadRequest, "Username, email, and password are required",
			"tags", []string{"REGISTER", "MISSING_REQUIRED_FIELDS"},
		)
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Step 4: Check username uniqueness constraint (prevents duplicate accounts)
	if err := db.Where("username = ?", registrationData.Username).First(&user.User{}).Error; err == nil {
		server.ResponseError(r.Context(),
			w, errors.New("username already taken"), http.StatusConflict, "Username already taken",
			"tags", []string{"REGISTER", "DB"},
		)
		return
	}

	// Step 5: Hash password using bcrypt with default cost (currently 10 rounds)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error generating hashed password",
			"tags", []string{"REGISTER", "PASSWORD"},
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 6: Construct user object with validated data and hashed password
	user := user.User{
		Username:     registrationData.Username,
		Email:        registrationData.Email,
		PasswordHash: string(hashedPassword),
		University:   registrationData.University,
		Language:     registrationData.Language,
	}

	// Persist new user to database
	if err := db.Create(&user).Error; err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating user",
			"tags", []string{"REGISTER", "DB"},
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert user struct to map for safe JSON response (removes sensitive fields)
	userMap := user.ToMap()

	// Step 7: Generate JWT tokens for immediate authentication after registration
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError,
			"Error getting session key",
			"tags", []string{"REGISTER", "SESSION_KEY"},
		)
		return
	}

	// Create short-lived access token (15 minutes) for API access
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating access token",
			"user_id", user.ID,
			"tags", []string{"REGISTER", "ACCESS_TOKEN"},
		)
		return
	}

	// Create long-lived refresh token (30 days) for token renewal
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating refresh token",
			"user_id", user.ID,
			"tags", []string{"REGISTER", "REFRESH_TOKEN"},
		)
		return
	}

	// Step 8: Cache user data in Redis for performance optimization (non-blocking)
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error marshalling user to json",
			"user_id", user.ID,
			"tags", []string{"REGISTER", "MARSHALLING"},
		)
		return
	}

	// Store in Redis hash with user ID as key (failure is non-critical, only logged)
	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(user.ID)), userJSON).Err(); err != nil {
		server.LogWarn(r.Context(),
			"Error caching user in redis", err,
			"tags", []string{"REGISTER", "REDIS"},
		)
	} else {
		server.LogInfo(r.Context(), "User cached successfully",
			"user_id", user.ID,
			"tags", []string{"REGISTER", "REDIS", "CACHED"},
		)
	}

	// Step 9: Send successful registration response with user data and tokens
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Register successful",
		"user":          userMap,
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	// Step 10: Log successful registration for audit trail and monitoring
	server.LogInfo(r.Context(), "User registered successfully",
		"user_id", user.ID,
		"username", user.Username,
		"tags", []string{"REGISTER", "WRITE"},
	)
}
