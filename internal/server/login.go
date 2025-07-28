package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"unipilot/internal/models/user"

	"github.com/spf13/viper"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
		PrintERROR(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	var user user.User
	if err := db.Where("username = ?", credentials.Username).First(&user).Error; err != nil {
		PrintERROR(w, http.StatusUnauthorized, "Invalid credentials")

		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {

		PrintERROR(w, http.StatusUnauthorized, "Invalid credentials")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create session
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error reading config file: %w", err))
	}

	SESSION_KEY := viper.GetString("SESSION_KEY")

	var store = sessions.NewCookieStore([]byte(SESSION_KEY))

	session, _ := store.Get(r, "session-auth")
	session.Values["user_id"] = user.ID
	session.Values["authenticated"] = true
	if err := session.Save(r, w); err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %w", err))
		return
	}

	id := strconv.Itoa(int(user.ID))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Login successful",
		"username": user.Username,
		"user_id":  id,
		"error":    "",
	})
}
