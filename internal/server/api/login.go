package server

import (
	"encoding/json"
	"net/http"
	"time"

	//"github.com/gorilla/sessions"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusBadRequest, "Invalid request body",
			"tags", []string{"LOGIN", "INVALID_REQUEST_BODY"},
		)
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	var user user.User
	if err := db.Where("username = ?", credentials.Username).First(&user).Error; err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusUnauthorized, "No user found",
			"username", credentials.Username,
			"tags", []string{"LOGIN", "USER", "DB"},
		)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusUnauthorized, "Invalid Password",
			"user_id", user.ID,
			"username", user.Username,
			"tags", []string{"LOGIN", "PASSWORD"},
		)
		return
	}

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Invalid session key:",
			"user_id", user.ID,
			"username", user.Username,
			"tags", []string{"LOGIN", "SESSION_KEY"},
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
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating access token",
			"user_id", user.ID,
			"username", user.Username,
			"tags", []string{"LOGIN", "ACCESS_TOKEN"},
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
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating refresh token",
			"user_id", user.ID,
			"username", user.Username,
			"tags", []string{"LOGIN", "REFRESH_TOKEN"},
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Login successful",
		"user":          user.ToMap(),
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	server.LogInfo(r.Context(),
		"Login successful",
		"user_id", user.ID,
		"username", user.Username,
		"tags", []string{"LOGIN"},
	)
}
