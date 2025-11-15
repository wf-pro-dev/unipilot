package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"unipilot/internal/server"
	grpc "unipilot/internal/server/api/grpc"
	"unipilot/internal/server/qdrant"
	"unipilot/internal/server/sse/grpc/notifications"
	"unipilot/internal/storage"

	"unipilot/internal/server/api/redis"

	Qdrant "github.com/qdrant/go-client/qdrant"
	Redis "github.com/redis/go-redis/v9"
)

var (
	GrpcClient   notifications.NotificationsServiceClient
	RedisClient  *Redis.Client
	QdrantClient *Qdrant.Client
)

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

	GrpcClient = *grpc.NewGRPCClient()
	defer grpc.CloseGRPCClient()

	RedisClient, err = redis.NewRedisClient()
	if err != nil {
		log.Println("Error getting redis client", err)
		return
	}
	defer RedisClient.Close()

	QdrantClient, err = qdrant.NewQdrantClient()
	if err != nil {
		log.Println("Error getting qdrant client", err)
		return
	}
	defer QdrantClient.Close()

	http.HandleFunc("/health", HealthHandler)

	http.HandleFunc(GetRouteName("register"), server.DBMiddleware(db, RegisterHandler))
	http.HandleFunc(GetRouteName("login"), server.DBMiddleware(db, LoginHandler))
	http.HandleFunc(GetRouteName("logout"), server.AuthMiddleware(LogoutHandler))

	http.HandleFunc(GetRouteName("token", "refresh"), server.AuthMiddleware(HandleRefreshToken))

	http.HandleFunc(GetRouteName("user"), server.DBMiddleware(db, server.AuthMiddleware(GetUserHandler)))
	http.HandleFunc(GetRouteName("user", "update"), server.DBMiddleware(db, server.AuthMiddleware(UpdateUserHandler)))

	http.HandleFunc(GetRouteName("users"), server.DBMiddleware(db, server.AuthMiddleware(GetUsersHandler)))

	http.HandleFunc(GetRouteName("assignment"), server.DBMiddleware(db, server.AuthMiddleware(CreateAssignmentHandler)))
	http.HandleFunc(GetRouteName("assignments"), server.DBMiddleware(db, server.AuthMiddleware(GetAssignmentHandler)))
	http.HandleFunc(GetRouteName("assignment", "update"), server.DBMiddleware(db, server.AuthMiddleware(UpdateAssignmentHandler)))

	http.HandleFunc(GetRouteName("course"), server.DBMiddleware(db, server.AuthMiddleware(CreateCourseHandler)))
	http.HandleFunc(GetRouteName("course", "get"), server.DBMiddleware(db, server.AuthMiddleware(GetCourseHandler)))
	http.HandleFunc(GetRouteName("course", "update"), server.DBMiddleware(db, server.AuthMiddleware(UpdateCourseHandler)))
	http.HandleFunc(GetRouteName("course", "link", "request"), server.DBMiddleware(db, server.AuthMiddleware(LinkRequestCourseHandler)))
	http.HandleFunc(GetRouteName("course", "link", "accept"), server.DBMiddleware(db, server.AuthMiddleware(AcceptLinkCourseHandler)))

	http.HandleFunc(GetRouteName("document"), server.DBMiddleware(db, server.AuthMiddleware(CreateDocumentHandler)))
	http.HandleFunc(GetRouteName("document", "download"), server.DBMiddleware(db, server.AuthMiddleware(DownloadDocumentHandler)))
	http.HandleFunc(GetRouteName("document", "delete"), server.DBMiddleware(db, server.AuthMiddleware(DeleteDocumentHandler)))
	http.HandleFunc(GetRouteName("document", "rag"), server.DBMiddleware(db, server.AuthMiddleware(UploadDocumentForRAGHandler)))

	http.HandleFunc(GetRouteName("note"), server.DBMiddleware(db, server.AuthMiddleware(CreateNoteHandler)))
	http.HandleFunc(GetRouteName("note", "get"), server.DBMiddleware(db, server.AuthMiddleware(GetNoteHandler)))
	http.HandleFunc(GetRouteName("note", "update"), server.DBMiddleware(db, server.AuthMiddleware(UpdateNoteHandler)))

	http.HandleFunc(GetRouteName("follow"), server.DBMiddleware(db, server.AuthMiddleware(HandleFollow)))
	http.HandleFunc(GetRouteName("followers"), server.DBMiddleware(db, server.AuthMiddleware(HandleGetFollowers)))
	http.HandleFunc(GetRouteName("following"), server.DBMiddleware(db, server.AuthMiddleware(HandleGetFollowing)))
	http.HandleFunc(GetRouteName("follow-status"), server.DBMiddleware(db, server.AuthMiddleware(HandleGetFollowStatus)))

	log.Println("Server listening on :3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
