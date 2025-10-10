package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"gorm.io/gorm"

	"unipilot/internal/secrets"
	"unipilot/internal/storage"
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

		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Register: %w", err))
			return
		}

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

func GetRouteName(name ...string) string {
	return fmt.Sprintf("/unipilot/api/v1/%s", strings.Join(name, "/"))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

func StartServer() {

	db, err := storage.GetRemoteDB()
	if err != nil {
		log.Println("Error getting database", err)
		return
	}

	sseServer = NewSSEServer()

	http.HandleFunc("/unipilot/events/v1", AuthMiddleware(sseServer.SSEHandler))
	http.HandleFunc("/health", HealthHandler)

	http.HandleFunc(GetRouteName("register"), DBMiddleware(db, RegisterHandler))
	http.HandleFunc(GetRouteName("login"), DBMiddleware(db, LoginHandler))
	http.HandleFunc(GetRouteName("logout"), AuthMiddleware(LogoutHandler))

	http.HandleFunc(GetRouteName("user"), DBMiddleware(db, AuthMiddleware(GetUserHandler)))
	http.HandleFunc(GetRouteName("user", "update"), DBMiddleware(db, AuthMiddleware(UpdateUserHandler)))

	http.HandleFunc(GetRouteName("users"), DBMiddleware(db, AuthMiddleware(GetUsersHandler)))

	http.HandleFunc(GetRouteName("assignment"), DBMiddleware(db, AuthMiddleware(CreateAssignmentHandler)))
	http.HandleFunc(GetRouteName("assignments"), DBMiddleware(db, AuthMiddleware(GetAssignmentHandler)))
	http.HandleFunc(GetRouteName("assignment", "update"), DBMiddleware(db, AuthMiddleware(UpdateAssignmentHandler)))

	http.HandleFunc(GetRouteName("course"), DBMiddleware(db, AuthMiddleware(CreateCourseHandler)))
	http.HandleFunc(GetRouteName("course", "get"), DBMiddleware(db, AuthMiddleware(GetCourseHandler)))
	http.HandleFunc(GetRouteName("course", "update"), DBMiddleware(db, AuthMiddleware(UpdateCourseHandler)))
	http.HandleFunc(GetRouteName("course", "link", "request"), DBMiddleware(db, AuthMiddleware(LinkRequestCourseHandler)))
	http.HandleFunc(GetRouteName("course", "link", "accept"), DBMiddleware(db, AuthMiddleware(AcceptLinkCourseHandler)))

	http.HandleFunc(GetRouteName("document"), DBMiddleware(db, AuthMiddleware(CreateDocumentHandler)))
	http.HandleFunc(GetRouteName("document", "download"), DBMiddleware(db, AuthMiddleware(DownloadDocumentHandler)))
	http.HandleFunc(GetRouteName("document", "delete"), DBMiddleware(db, AuthMiddleware(DeleteDocumentHandler)))

	http.HandleFunc(GetRouteName("note"), DBMiddleware(db, AuthMiddleware(CreateNoteHandler)))
	http.HandleFunc(GetRouteName("note", "get"), DBMiddleware(db, AuthMiddleware(GetNoteHandler)))
	http.HandleFunc(GetRouteName("note", "update"), DBMiddleware(db, AuthMiddleware(UpdateNoteHandler)))

	http.HandleFunc(GetRouteName("follow"), DBMiddleware(db, AuthMiddleware(HandleFollow)))
	http.HandleFunc(GetRouteName("followers"), DBMiddleware(db, AuthMiddleware(HandleGetFollowers)))
	http.HandleFunc(GetRouteName("following"), DBMiddleware(db, AuthMiddleware(HandleGetFollowing)))
	http.HandleFunc(GetRouteName("follow-status"), DBMiddleware(db, AuthMiddleware(HandleGetFollowStatus)))

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
