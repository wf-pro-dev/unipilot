package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/server"

	"unipilot/internal/storage"

	"unipilot/internal/errors"
)

func HealthHandler(c *fiber.Ctx) error {
	err := c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
	return err
}

func StartServer() error {
	// Initialize logger first so we can use proper logging for startup errors
	server.InitLogger()

	// Initialize Fiber app
	app := fiber.New()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "component", "system")

	db, err := storage.GetRemoteDB()
	if err != nil {
		return errors.Wrap(err, errors.DBConnectionFailed, "failed to initialize database connection")
	}

	err = NewGRPCClient()
	if err != nil {
		return errors.Wrap(err, errors.GRPCFailed, "cannot connect to sse gRPC")
	}
	defer func() {
		if err := CloseGRPC(); err != nil {
			server.LogError(context.Background(), errors.WrapServer(err, errors.GRPCCloseFailed, "Failed to close gRPC client", fiber.StatusInternalServerError))
		}
	}()

	err = NewRedisClient()
	if err != nil {
		return errors.Wrap(err, errors.RedisFailed, "cannot connect to redis")
	}
	defer func() {
		if err := CloseRedis(); err != nil {
			server.LogError(context.Background(), errors.WrapServer(err, errors.RedisCloseFailed, "Failed to close redis client", fiber.StatusInternalServerError))
		}
	}()

	err = NewQdrantClient()
	if err != nil {
		return errors.Wrap(err, errors.QdrantFailed, "cannot connect to qdrant")
	}
	defer func() {
		if err := CloseQdrant(); err != nil {
			server.LogError(context.Background(), errors.WrapServer(err, errors.QdrantCloseFailed, "Failed to close qdrant client", fiber.StatusInternalServerError))
		}
	}()

	app.Get("/health", HealthHandler)

	app.Use(server.LoggerMiddleware)
	app.Use(server.DBMiddleware(db))
	app.Use(server.ErrorHandlerMiddleware)

	app.Post("/auth/register", RegisterHandler)
	app.Post("/auth/login", LoginHandler)

	app.Use(server.AuthMiddleware)
	// Protected routes
	app.Post("/auth/logout", LogoutHandler)
	app.Post("/auth/refresh-token", RefreshTokenHandler)

	app.Get("/users", GetUsersHandler)
	app.Get("/users/me", GetUserHandler)
	app.Post("/users/me", UpdateUserHandler)
	app.Post("/users/me/profile-picture", UpdateProfilePictureHandler)

	app.Post("/users/:id/follow", HandleFollow)
	app.Get("/users/:id/followers", HandleGetFollowers)
	app.Get("/users/:id/following", HandleGetFollowing)
	app.Get("/assignments", GetAssignmentHandler)
	app.Post("/assignments", CreateAssignmentHandler)
	app.Put("/assignments/:id", UpdateAssignmentHandler)
	app.Delete("/assignments/:id", DeleteAssignmentHandler)

	app.Get("/courses", GetCoursesHandler)
	app.Post("/courses", CreateCourseHandler)
	app.Put("/courses/:id", UpdateCourseHandler)
	app.Post("/courses/:id/link-request", LinkRequestCourseHandler)
	app.Post("/courses/:id/link-accept", AcceptLinkCourseHandler)
	app.Delete("/courses/:id", DeleteCourseHandler)

	app.Get("/documents/assignments/:id", GetAssignmentDocumentsHandler)
	app.Get("/documents/assignments/:id/rag", GetAssignmentDocumentIDsRAG)

	app.Get("/documents", GetDocumentsHandler)
	app.Post("/documents", CreateDocumentHandler)
	app.Post("/documents/:id/download", DownloadDocumentHandler)
	app.Delete("/documents/:id", DeleteDocumentHandler)
	app.Post("/documents/:id/rag", UploadDocumentForRAGHandler)
	app.Delete("/documents/:id/rag", DeleteDocumentRAG)

	app.Get("/notes", GetNotesHandler)
	app.Post("/notes", CreateNoteHandler)
	app.Post("/notes/stream", CreateNoteStreamHandler)
	app.Put("/notes/:id", UpdateNoteHandler)
	app.Delete("/notes/:id", DeleteNoteHandler)

	if err := app.Listen(":3000"); err != nil {
		return errors.Wrap(err, errors.APIStartFailed, "cannot start api server")
	}
	return nil
}
