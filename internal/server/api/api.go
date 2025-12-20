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

	// Protected routes
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
	app.Post("/courses/link-accept", server.AuthMiddleware, AcceptLinkCourseHandler)
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
	app.Post("/notes/stream", server.AuthMiddleware, CreateNoteStreamHandler)
	app.Put("/notes/:id", server.AuthMiddleware, UpdateNoteHandler)
	app.Delete("/notes/:id", server.AuthMiddleware, DeleteNoteHandler)

	if err := app.Listen(":3000"); err != nil {
		return errors.Wrap(err, errors.APIStartFailed, "cannot start api server")
	}
	return nil
}
