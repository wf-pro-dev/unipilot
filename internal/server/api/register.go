package server

import (
	"context"
	"encoding/json"
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

	var registrationData struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Password   string `json:"password"`
		University string `json:"university"`
		Language   string `json:"language"`
	}

	err := json.NewDecoder(r.Body).Decode(&registrationData)
	if err != nil {
		server.PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	// Validate input
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.University == "" || registrationData.Language == "" {
		server.PrintERROR(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, "Could not process password")
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
		server.PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating user: %s", err))
		return
	}

	userMap := user.ToMap()

	// Create session
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Register: %s", err.Error()))
		return
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Register: %s", err.Error()))
		return
	}

	// Cache the new user in redis
	userJSON, err := json.Marshal(userMap)
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling user to json: %v", err))
		return
	}

	if err := RedisClient.HSet(context.Background(), "users", strconv.Itoa(int(user.ID)), userJSON).Err(); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error caching user in redis: %v", err))
		return
	}

	server.PrintLOG([]string{"INFO", "USER", "REGISTER", "REDIS"}, fmt.Sprintf("User cached successfully for user id: %d", user.ID))

	/* DEPRECATED

	var store = sessions.NewCookieStore([]byte(SESSION_KEY))

	session, _ := store.Get(r, "session-auth")
	session.Values["user_id"] = user.ID
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %w", err))
		return
	}
	*/

	id := strconv.Itoa(int(user.ID))

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"id":      id,
		"user":    userMap,
		"token":   token,
	})

	server.PrintLOG([]string{"SUCCESS", "USER", "REGISTER"}, fmt.Sprintf("User ID : %v, Username : %v", id, user.Username))
}
