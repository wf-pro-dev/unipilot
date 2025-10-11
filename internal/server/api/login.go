package server

import (
	"encoding/json"
	"fmt"
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
		server.PrintERROR(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	var user user.User
	if err := db.Where("username = ?", credentials.Username).First(&user).Error; err != nil {
		server.PrintERROR(w, http.StatusUnauthorized, "Invalid credentials")

		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {

		server.PrintERROR(w, http.StatusUnauthorized, "Invalid credentials")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Login: %w", err))
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
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Login: %w", err))
		return
	}

	/* DEPRECATED
	var store = sessions.NewCookieStore([]byte(SESSION_KEY))

	session, _ := store.Get(r, "session-auth")
	session.Values["user_id"] = user.ID
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %w", err))
		return
	} */

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user":    user.ToMap(),
		"token":   token,
	})

	server.PrintLOG([]string{"SUCCESS", "LOGIN"}, fmt.Sprintf("User ID : %v, Username : %v", user.ID, user.Username))
}
