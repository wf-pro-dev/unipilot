package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"

	"unipilot/internal/storage"
	
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

	http.HandleFunc("/acc-homework/events", AuthMiddleware(sseServer.SSEHandler))

	http.HandleFunc("/acc-homework/register", DBMiddleware(db, RegisterHandler))
	http.HandleFunc("/acc-homework/login", DBMiddleware(db, LoginHandler))
	http.HandleFunc("/acc-homework/logout", AuthMiddleware(LogoutHandler))
	http.HandleFunc("/acc-homework/user", DBMiddleware(db, AuthMiddleware(GetUserHandler)))

	http.HandleFunc("/acc-homework/assignment", DBMiddleware(db, AuthMiddleware(CreateAssignmentHandler)))
	http.HandleFunc("/acc-homework/assignment/get", DBMiddleware(db, AuthMiddleware(GetAssignmentHandler)))
	http.HandleFunc("/acc-homework/assignment/update", DBMiddleware(db, AuthMiddleware(UpdateAssignmentHandler)))

	http.HandleFunc("/acc-homework/course", DBMiddleware(db, AuthMiddleware(CreateCourseHandler)))
	http.HandleFunc("/acc-homework/course/get", DBMiddleware(db, AuthMiddleware(GetCourseHandler)))
	http.HandleFunc("/acc-homework/course/update", DBMiddleware(db, AuthMiddleware(UpdateCourseHandler)))
	
	http.HandleFunc("/acc-homework/document/metadata", DBMiddleware(db, AuthMiddleware(CreateDocumentMetadataHandler)))
	http.HandleFunc("/acc-homework/document/metadata/delete", DBMiddleware(db, AuthMiddleware(DeleteDocumentMetadataHandler)))

	http.HandleFunc("/acc-homework/note", DBMiddleware(db, AuthMiddleware(CreateNoteHandler)))
	http.HandleFunc("/acc-homework/note/get", DBMiddleware(db, AuthMiddleware(GetNoteHandler)))
	http.HandleFunc("/acc-homework/note/update", DBMiddleware(db, AuthMiddleware(UpdateNoteHandler)))
	
	log.Println("Server listening on :3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}


func PrintLog(message string) {
	log.Printf("[INFO] %s", message)
}

func PrintERROR(w http.ResponseWriter, code int, message string) {
	log.Printf("[ERROR] [%d] %s", code, message)
	http.Error(w, message, code)
}

