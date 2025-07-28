package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/viper"

	"unipilot/internal/models/user"
	"unipilot/internal/services/notion"

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
		NotionSecret string `json:"notion_secret"`
		NotionPageID string `json:"notion_page_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&registrationData)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	db := r.Context().Value("db").(*gorm.DB)

	// Validate input
	if registrationData.Username == "" || registrationData.Email == "" || registrationData.Password == "" || registrationData.NotionSecret == "" || registrationData.NotionPageID == "" {
		PrintERROR(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	databases_id, err := getNotionDB(registrationData.NotionPageID, registrationData.NotionSecret)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("error fetching notion databases id %s", err))
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
		NotionAPIKey:    registrationData.NotionSecret,
		AssignmentsDbId: databases_id["Assignments"],
		CoursesDbId:     databases_id["Courses"],
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

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func getNotionDB(page_id, user_api_key string) (map[string]string, error) {

	url := fmt.Sprintf("blocks/%s/children?page_size=100", page_id)
	respBody, err := notion.SendNotionRequest(nil, "GET", url, user_api_key)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	type NotionDB struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		ChildDatabase struct {
			Title string `json:"title"`
		} `json:"child_database"`
	}

	type Response struct {
		Results []NotionDB `json:"results"`
	}

	var notionResp Response
	if err := json.Unmarshal(respBody, &notionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var database_id = make(map[string]string)

	for _, resp := range notionResp.Results {
		if resp.Type == "child_database" {
			database_id[resp.ChildDatabase.Title] = resp.ID
		}

	}

	return database_id, nil

}
