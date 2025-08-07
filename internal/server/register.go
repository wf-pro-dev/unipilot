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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var registrationData struct {
		Username     string `json:"username"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		University   string `json:"university"`
		Language     string `json:"language"`

	}

	err := json.NewDecoder(r.Body).Decode(&registrationData)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	// Validate input
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.University == "" || registrationData.Language == "" {
		PrintERROR(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registrationData.Password), bcrypt.DefaultCost)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Could not process password")
		return
	}

	// Create user
	user := user.User{
		Username:        registrationData.Username,
		Email:           registrationData.Email,
		PasswordHash:    string(hashedPassword),
		University:	 registrationData.University,
		Language:	 registrationData.Language,
	}

	result := db.Create(&user)
	if result.Error != nil {
		PrintERROR(w, http.StatusConflict, "Username or email already exists")
		return
	}

	// Create session
	viper.SetConfigFile(".env")
	err = viper.ReadInConfig()
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
	
	PrintLog(fmt.Sprintf("%v",id))
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"id" : id,
		"user": user.ToMap(),
	})
}
