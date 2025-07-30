package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"

	"unipilot/internal/models/user"
	"unipilot/internal/storage"
	"unipilot/internal/types"

	"github.com/gorilla/sessions"

	"gorm.io/gorm"
)

var sseServer *SSEServer

// MiddleWares ! put on separate file

func DBMiddleware(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next(w, r.WithContext(ctx))
	}
}

// AuthMiddleware checks if the user is authenticated
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		viper.SetConfigFile(".env")
		err := viper.ReadInConfig()
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error reading config file: %w", err))
			return
		}

		SESSION_KEY := viper.GetString("SESSION_KEY")

		var store = sessions.NewCookieStore([]byte(SESSION_KEY))

		session, err := store.Get(r, "session-auth")
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %w", err))
			return
		}

		// Check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			PrintERROR(w, http.StatusUnauthorized, "Unauthorized - please login")
			return
		}

		// You can also add the user ID to the request context if needed
		userID, ok := session.Values["user_id"].(uint)
		if ok {
			ctx := context.WithValue(r.Context(), "user_id", userID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	}
}

func StartServer() {

	db, err := storage.GetRemoteDB()
	if err != nil {
		log.Println("Error getting database", err)
		return
	}

	sseServer = NewSSEServer(db)

	http.HandleFunc("/webhooks", DBMiddleware(db, notionWebhookHandler))
	http.HandleFunc("/acc-homework/events", AuthMiddleware(sseServer.SSEHandler))

	http.HandleFunc("/acc-homework/register", DBMiddleware(db, RegisterHandler))
	http.HandleFunc("/acc-homework/login", DBMiddleware(db, LoginHandler))
	http.HandleFunc("/acc-homework/logout", AuthMiddleware(LogoutHandler))
	http.HandleFunc("/acc-homework/user", DBMiddleware(db, AuthMiddleware(GetUserHandler)))

	http.HandleFunc("/acc-homework/assignment", DBMiddleware(db, AuthMiddleware(CreateAssignmentHandler)))
	http.HandleFunc("/acc-homework/assignment/get", DBMiddleware(db, AuthMiddleware(GetAssignmentHandler)))
	http.HandleFunc("/acc-homework/assignment/update", DBMiddleware(db, AuthMiddleware(UpdateAssignmentHandler)))

	//http.HandleFunc("/acc-homework/course", DBMiddleware(db, AuthMiddleware(CreateCourseHandler)))
	http.HandleFunc("/acc-homework/course/get", DBMiddleware(db, AuthMiddleware(GetCourseHandler)))
	http.HandleFunc("/acc-homework/course/update", DBMiddleware(db, AuthMiddleware(UpdateCourseHandler)))

	http.HandleFunc("/acc-homework/document/metadata", DBMiddleware(db, AuthMiddleware(CreateDocumentMetadataHandler)))
	http.HandleFunc("/acc-homework/document/metadata/delete", DBMiddleware(db, AuthMiddleware(DeleteDocumentMetadataHandler)))

	http.HandleFunc("/notion-webhooks/test", testHandler)
	log.Println("Server listening on :3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func webhookTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Token string `json:"verification_token"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("%s Token: %s", time.Now().Format(time.Stamp), payload.Token)
}

func testHandler(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s test: %s", time.Now().Format(time.Stamp), r.URL.Path)

	var payload struct {
		Test string `json:"test"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("%s test: %s", time.Now().Format(time.Stamp), payload.Test)
}

func PrintLog(message string) {
	log.Printf("[INFO] %s", message)
}

func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}

func notionWebhookHandler(w http.ResponseWriter, r *http.Request) {

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	// 1. Verify it's a POST request
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Decode the payload
	var payload types.NotionWebhookPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error decoding payload: %s", err))
		return
	}

	// 3. Get the author of the change
	var author_type string
	if len(payload.Authors) > 0 {
		author_type = payload.Authors[0].Type
		if author_type == "bot" {
			PrintLog("Author is a bot, skipping")
			return
		}
	}

	user_notion_id := payload.Authors[0].Id

	u, err := user.Get_User_by_NotionID(user_notion_id, db)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting user: %s", err))
	}

	// 4. Handle the payload
	switch payload.Type {
	case "page.properties_updated":
		WebhookUpdateHandler(w, r, payload, u)
	case "page.created":
		WebhookCreateHandler(w, r, payload, u)
	case "page.deleted":
		WebhookDeleteHandler(w, r, payload)
	default:
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Unknown payload type: %s", payload.Type))
	}
}
