package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := r.Context().Value("request_id").(string)
	startTime := r.Context().Value("start_time").(time.Time)

	var registrationData struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		University string `json:"university"`
		Language   string `json:"language"`
	}

	err := json.NewDecoder(r.Body).Decode(&registrationData)
	if err != nil {
		server.ResponseError(
			w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER"},
		)

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	// Validate input
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.University == "" || registrationData.Language == "" {
		server.ResponseError(
			w, err, http.StatusBadRequest, "Username, email, and password are required",
			"request_id", requestID,
			"Register: Username, email, and password are required",
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "MISSING_REQUIRED_FIELDS"},
		)

		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)

		return
	}

	// Check if username is already taken
	if err := db.Where("username = ?", registrationData.Username).First(&user.User{}).Error; err == nil {
		server.ResponseError(
			w, errors.New("username already taken"), http.StatusConflict, "Username already taken",
			"request_id", requestID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "DB"},
		)

		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	if err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError, "Error generating hashed password",
			"request_id", requestID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "PASSWORD"},
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	// Create user
	user := user.User{
		Username:     registrationData.Username,
		Email:        registrationData.Email,
		PasswordHash: string(hashedPassword),
		University:   registrationData.University,
		Language:     registrationData.Language,
	}

	if err := db.Create(&user).Error; err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError, "Error creating user",
			"request_id", requestID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "DB"},
		)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	userMap := user.ToMap()

	// Create session
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError,
			"Error getting session key",
			"request_id", requestID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "SESSION_KEY"},
		)
		return
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError, "Error creating access token",
			"request_id", requestID,
			"user_id", user.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "ACCESS_TOKEN"},
		)

		return
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError, "Error creating refresh token",
			"request_id", requestID,
			"user_id", user.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "REFRESH_TOKEN"},
		)

		return
	}

	// Cache the new user in redis
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		server.ResponseError(
			w, err, http.StatusInternalServerError, "Error marshalling user to json",
			"request_id", requestID,
			"user_id", user.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "MARSHALLING"},
		)
		return
	}

	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(user.ID)), userJSON).Err(); err != nil {

		server.LogWarn(
			"Error caching user in redis", err,
			"request_id", requestID,
			"user_id", user.ID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"REGISTER", "REDIS"},
		)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Register successful",
		"user":          userMap,
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	server.LogInfo(fmt.Sprintf(" User registered successfully user id: %d, username: %s", user.ID, user.Username),
		"request_id", requestID,
		"user_id", user.ID,
		"username", user.Username,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"REGISTER", "WRITE"},
	)
}
