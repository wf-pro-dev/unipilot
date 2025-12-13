package server

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

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

func HealthHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func StartServer() {
	// Initialize logger first so we can use proper logging for startup errors
	server.InitLogger()

	// Initialize Fiber app
	app := fiber.New()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "component", "system")

	db, err := storage.GetRemoteDB()
	if err != nil {
		server.LogFatal(ctx, "Failed to initialize database connection", err,
			"tags", []string{"system", "db", "high"},
			"error_type", "database",
		)
		return
	}

	GrpcClient = *grpc.NewGRPCClient()
	defer grpc.CloseGRPCClient()

	RedisClient, err = redis.NewRedisClient()
	if err != nil {
		server.LogFatal(ctx, "Failed to initialize Redis client", err,
			"tags", []string{"cache", "network", "high"},
			"error_type", "network",
		)
		return
	}
	defer RedisClient.Close()

	QdrantClient, err = qdrant.NewQdrantClient()
	if err != nil {
		server.LogFatal(ctx, "Failed to initialize Qdrant client", err,
			"tags", []string{"rag", "network", "high"},
			"error_type", "network",
		)
		return
	}
	defer QdrantClient.Close()

	app.Get("/health", HealthHandler)

	app.Use(server.LoggerMiddleware)
	app.Use(server.DBMiddleware(db))

	app.Post("/auth/register", RegisterHandler)
	app.Post("/auth/login", LoginHandler)
	app.Post("/auth/logout", server.AuthMiddleware, LogoutHandler)
	app.Post("/auth/refresh-token", server.AuthMiddleware, RefreshTokenHandler)

	app.Get("/users", server.AuthMiddleware, GetUsersHandler)
	app.Get("/users/me", server.AuthMiddleware, GetUserHandler)
	app.Post("/users/me", server.AuthMiddleware, UpdateUserHandler)
	app.Post("/users/me/profile-picture", server.AuthMiddleware, UpdateProfilePictureHandler)

	app.Post("/users/:id/follow", server.AuthMiddleware, HandleFollow)
	app.Get("/users/:id/followers", server.AuthMiddleware, HandleGetFollowers)
	app.Get("/users/:id/following", server.AuthMiddleware, HandleGetFollowing)

	app.Get("/assignments", server.AuthMiddleware, GetAssignmentHandler)
	app.Post("/assignments", server.AuthMiddleware, CreateAssignmentHandler)
	app.Put("/assignments/:id", server.AuthMiddleware, UpdateAssignmentHandler)
	app.Delete("/assignments/:id", server.AuthMiddleware, DeleteAssignmentHandler)

	app.Get("/courses", server.AuthMiddleware, GetCoursesHandler)
	app.Post("/courses", server.AuthMiddleware, CreateCourseHandler)
	app.Put("/courses/:id", server.AuthMiddleware, UpdateCourseHandler)
	app.Post("/courses/:id/link-request", server.AuthMiddleware, LinkRequestCourseHandler)
	app.Post("/courses/:id/link-accept", server.AuthMiddleware, AcceptLinkCourseHandler)
	app.Delete("/courses/:id", server.AuthMiddleware, DeleteCourseHandler)

	app.Get("/documents/assignments/:id", server.AuthMiddleware, GetAssignmentDocumentsHandler)
	app.Get("/documents/assignments/:id/rag", server.AuthMiddleware, GetAssignmentDocumentIDsRAG)

	app.Get("/documents", server.AuthMiddleware, GetDocumentsHandler)
	app.Post("/documents", server.AuthMiddleware, CreateDocumentHandler)
	app.Post("/documents/:id/download", server.AuthMiddleware, DownloadDocumentHandler)
	app.Delete("/documents/:id", server.AuthMiddleware, DeleteDocumentHandler)
	app.Post("/documents/:id/rag", server.AuthMiddleware, UploadDocumentForRAGHandler)
	app.Delete("/documents/:id/rag", server.AuthMiddleware, DeleteDocumentRAG)

	app.Get("/notes", server.AuthMiddleware, GetNotesHandler)
	app.Post("/notes", server.AuthMiddleware, CreateNoteHandler)
	app.Put("/notes/:id", server.AuthMiddleware, UpdateNoteHandler)
	app.Delete("/notes/:id", server.AuthMiddleware, DeleteNoteHandler)

	log.Fatal(app.Listen(":3000"))
}
