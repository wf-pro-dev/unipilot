package server

import (
	"context"
	Errors "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RouteTags defines tags for each route pattern
// Format: [resource, operation, level, action?]
// - resource: user, assignment, course, document, note, follow, auth, system
// - operation: db, io, network, auth
// - level: high (>1s), medium (100ms-1s), low (<100ms)
// - action: create, read, update, delete (for CRUD operations)
//
// Key format: "METHOD /route/pattern" (e.g., "GET /assignments", "PUT /assignments/:id")
// This ensures routes are uniquely identified by HTTP method and route pattern.
//
// Consolidated taxonomy:
// - Merged: upload/download/storage → io (file operations)
// - Merged: rag → document (all document-related operations)
// - Merged: compute → db (database operations for RAG metadata)
// - Merged: network → db (course link is primarily DB operation with network side-effect)
var routeTags = map[string][]string{
	// Authentication routes
	"POST /auth/register":      {"user", "db", "high", "create"},
	"POST /auth/login":         {"login", "auth", "medium", "read"},
	"POST /auth/logout":        {"logout", "auth", "low"},
	"POST /auth/refresh-token": {"token", "auth", "medium"},

	// User routes
	"GET /users/me":                  {"user", "db", "low", "read"},
	"POST /users/me":                 {"user", "db", "medium", "update"},
	"POST /users/me/profile-picture": {"user", "storage", "high", "update"},
	"GET /users":                     {"user", "db", "medium", "read"},
	"POST /users/:id/follow":         {"follow", "db", "medium", "update"},
	"GET /users/:id/followers":       {"follow", "db", "low", "read"},
	"GET /users/:id/following":       {"follow", "db", "low", "read"},

	// Assignment routes
	"GET /assignments":        {"assignment", "db", "low", "read"},
	"POST /assignments":       {"assignment", "db", "medium", "create"},
	"PUT /assignments/:id":    {"assignment", "db", "medium", "update"},
	"DELETE /assignments/:id": {"assignment", "db", "medium", "delete"},

	// Course routes
	"GET /courses":                   {"course", "db", "low", "read"},
	"POST /courses":                  {"course", "db", "medium", "create"},
	"PUT /courses/:id":               {"course", "db", "medium", "update"},
	"DELETE /courses/:id":            {"course", "db", "medium", "delete"},
	"POST /courses/:id/link-request": {"course", "db", "medium"},
	"POST /courses/:id/link-accept":  {"course", "db", "high"},

	// Note routes
	"GET /notes":         {"note", "db", "low", "read"},
	"POST /notes":        {"note", "db", "medium", "create"},
	"POST /notes/stream": {"note", "gemini", "high", "stream"},
	"PUT /notes/:id":     {"note", "db", "medium", "update"},
	"DELETE /notes/:id":  {"note", "db", "medium", "delete"},

	// Document routes (includes RAG operations)
	"GET /documents":                     {"document", "db", "low", "read"},
	"POST /documents":                    {"document", "storage", "high", "create"},
	"DELETE /documents/:id":              {"document", "storage", "medium", "delete"},
	"POST /documents/:id/download":       {"document", "storage", "medium"},
	"POST /documents/:id/rag":            {"document", "rag", "high", "create"},
	"DELETE /documents/:id/rag":          {"document", "rag", "medium", "delete"},
	"GET /documents/assignments/:id":     {"document", "db", "low", "read"},
	"GET /documents/assignments/:id/rag": {"document", "rag", "low", "read"},

	// Follow routes (also under users, but keeping for completeness)
	"GET /follow-status": {"follow", "db", "low", "read"},
}

// getRouteTags returns tags for a given HTTP method and route pattern
// Uses "METHOD /route/pattern" format to uniquely identify routes
// Falls back to generic tags if route not found
func getRouteTags(method, routePattern string) []string {
	// Normalize method to uppercase
	method = strings.ToUpper(method)

	// Remove query parameters from route pattern if present
	cleanPattern := routePattern
	if idx := strings.Index(routePattern, "?"); idx != -1 {
		cleanPattern = routePattern[:idx]
	}

	// Create composite key: "METHOD /route/pattern"
	key := method + " " + cleanPattern

	// Try exact match first
	if tags, ok := routeTags[key]; ok {
		return tags
	}

	// Fallback to generic tags
	return []string{"system", "network", "medium"}
}

// getClientIP extracts the real client IP from request headers
// Handles Docker Swarm, nginx, and other proxy scenarios
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header (most common, contains chain of IPs)
	// Format: "client, proxy1, proxy2" - we want the first (original client)
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Fallback to X-Real-IP (nginx sets this)
	if realIP := c.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Final fallback to RemoteAddr (direct connection or no proxy headers)
	// Remove port if present
	addr := c.IP()
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		addr = addr[:idx]
	}
	return addr
}

// DBMiddleware injects the database connection into Fiber's locals.
// Provides database access to all downstream handlers without requiring explicit
// parameter passing. Essential middleware for all database-dependent endpoints.
//
// Parameters:
//   - db: GORM database connection instance
//
// Returns:
//   - fiber.Handler: Middleware function that adds database to Fiber locals
//
// Locals Added:
//   - "db": *gorm.DB instance for database operations
//
// Usage:
//   - Applied to all API routes that require database access
//   - Handlers can retrieve database via: c.Locals("db").(*gorm.DB)
//
// Security Considerations:
//   - Database connection is shared across requests (connection pooling)
//   - No user-specific database isolation (handled at application level)
func DBMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	}
}

// LoggerMiddleware provides comprehensive request logging and performance tracking.
// Generates unique request IDs, tracks request duration, and logs detailed request
// information for debugging, monitoring, and audit purposes.
//
// Returns:
//   - fiber.Handler: Middleware function that adds logging context and tracks requests
//
// Locals Added:
//   - "request_id": Unique UUID for request tracing across services
//   - "start_time": Request start time for duration calculation
//
// Logging Features:
//   - Unique request ID generation for distributed tracing
//   - Request duration tracking for performance monitoring
//   - Comprehensive request details (method, path, remote address)
//   - User ID tracking when available from authentication context
//   - Dual logging (debug console + structured file logging)
//
// Performance Metrics:
//   - Request duration in milliseconds
//   - Request volume tracking
//   - User activity monitoring
//
// Security Features:
//   - Request tracing for security audit trails
//   - User activity logging for compliance
//   - Remote address tracking for security analysis
func LoggerMiddleware(c *fiber.Ctx) error {
	// Generate unique request ID and capture start time for tracking
	startTime := time.Now()
	requestID := uuid.New().String()

	// Get the actual request path (with parameters, e.g., "/assignments/289")
	// This is what we'll log in the output
	requestPath := c.Path()

	// Store in locals for handlers
	c.Locals("request_id", requestID)
	c.Locals("start_time", startTime)
	c.Locals("component", "api")

	// Execute the next handler (route matching happens here)
	err := c.Next()

	if errorHandled, ok := c.Locals("error_handled").(bool); ok && errorHandled {
		// Error was already logged and handled, don't log again
		return err
	}

	// Get the route pattern from Fiber after route matching (e.g., "/assignments/:id")
	// This is what we use for tag matching
	routePattern := requestPath
	if route := c.Route(); route != nil && route.Path != "" {
		routePattern = route.Path
	}

	// Get route-specific tags using method + route pattern
	tags := getRouteTags(c.Method(), routePattern)

	// Calculate request duration
	duration := time.Since(startTime).Milliseconds()

	// Get real client IP (handles Docker Swarm VIP issue)
	clientIP := getClientIP(c)

	// Create context for logging
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", requestID)
	ctx = context.WithValue(ctx, "start_time", startTime)
	ctx = context.WithValue(ctx, "status_code", c.Response().StatusCode())
	ctx = context.WithValue(ctx, "duration", duration)

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		userID = 0
	}
	ctx = context.WithValue(ctx, "user_id", userID)

	message, ok := c.Locals("message").(string)
	if !ok {
		message = "Request completed"
		LogWarn(ctx, errors.WrapServer(fmt.Errorf("route message not found"), errors.ValidationInvalid, "Route message not found", fiber.StatusInternalServerError))
	}

	logLevel := "INFO"
	if duration > 1000 {
		logLevel = "WARN"
		// Update tags level to "high" for slow requests
		if len(tags) >= 3 {
			tags[2] = "high"
		}
	}

	// Log slow request
	if logLevel == "WARN" && err == nil {
		LogWarn(ctx, errors.WrapServer(fmt.Errorf("slow request: %dms", duration), errors.SlowRequest, "Slow request", fiber.StatusInternalServerError))
	} else {
		LogInfo(ctx, message,
			"method", c.Method(),
			"path", requestPath,
			"remote_addr", clientIP,
			"tags", tags,
		)
	}

	return err
}

// Claims represents the JWT token payload structure for user authentication.
// Embeds the complete user object and standard JWT claims for comprehensive
// authentication and authorization. Used for stateless authentication across
// all protected API endpoints.
//
// Fields:
//   - User: Complete user object with all profile information
//   - jwt.RegisteredClaims: Standard JWT claims (issued at, expires at, etc.)
//
// Security Features:
//   - Contains full user context to avoid additional database lookups
//   - Includes standard JWT expiration and validation claims
//   - Signed with server secret key for tamper protection
//
// Token Lifespans:
//   - Access tokens: 15 minutes (short-lived for security)
//   - Refresh tokens: 30 days (long-lived for user convenience)
type Claims struct {
	User                 user.User `json:"user"` // Complete user object for context
	jwt.RegisteredClaims           // Standard JWT claims (exp, iat, etc.)
}

// AuthMiddleware provides JWT-based authentication for protected API endpoints.
// Validates JWT tokens from Authorization headers, extracts user context, and
// injects user information into request context for downstream handlers.
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.HandlerFunc: Middleware function that validates authentication
//
// Authentication Flow:
//   1. Extracts JWT token from "Authorization: Bearer <token>" header
//   2. Validates token signature using server secret key
//   3. Checks token expiration and claims validity
//   4. Extracts user object from token claims
//   5. Injects user context into request for downstream handlers
//
// Context Values Added:
//   - "user": Complete user.User object from JWT claims
//   - "user_id": User ID (uint) for convenient access
//
// Security Features:
//   - JWT signature validation prevents token tampering
//   - Token expiration enforcement for security
//   - Stateless authentication (no server-side sessions)
//   - Automatic user context injection for authorization
//
// Error Handling:
//   - Missing Authorization header: Silent failure (allows public endpoints)
//   - Invalid token format: Silent failure (malformed Bearer token)
//   - Invalid/expired token: HTTP 401 Unauthorized response
//   - Invalid claims: HTTP 401 Unauthorized response
//
// Usage:
//   - Applied to all protected API routes
//   - Handlers access user via: r.Context().Value("user").(user.User)
//   - User ID available via: r.Context().Value("user_id").(uint)

func AuthMiddleware(c *fiber.Ctx) error {
	// Step 1: Extract JWT token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errors.WrapServer(
			fmt.Errorf("authorization header required"),
			errors.AuthUnauthorized,
			"Authorization header required",
			fiber.StatusUnauthorized,
		)
	}

	// Step 2: Parse Bearer token format ("Bearer <token>")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return errors.WrapServer(
			fmt.Errorf("authorization Token invalid"),
			errors.AuthTokenInvalid,
			"Authorization Token invalid",
			fiber.StatusUnauthorized,
		)
	}
	// Step 3: Retrieve server secret key for token validation
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		return errors.Inherit(err, errors.ConfigEnvVarNotFound).ToServerError(fiber.StatusInternalServerError)
	}
	// Step 4: Parse and validate JWT token with server secret
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SESSION_KEY), nil // Use same secret as token generation
	})
	// Step 5: Validate token parsing and signature
	if err != nil || !token.Valid {
		return errors.WrapServer(
			err,
			errors.AuthTokenInvalid,
			"Invalid or expired token",
			fiber.StatusUnauthorized,
		)
	}
	// Step 6: Extract and validate claims structure
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return errors.WrapServer(
			fmt.Errorf("invalid token claims"),
			errors.AuthTokenInvalid,
			"Invalid token claims",
			fiber.StatusUnauthorized,
		)
	}
	// Step 7: Extract user object from validated claims
	user := claims.User

	// Step 8: Inject user context into Fiber locals for downstream handlers
	c.Locals("user", user)
	c.Locals("user_id", user.ID)

	// Step 9: Continue to next handler with authenticated user context
	return c.Next()
}

// DEPRECATED: This middleware is deprecated in favor of JWT-based authentication (AuthMiddleware).
// AuthMiddlewareV1 provides legacy session-based authentication using cookies.
// This is the original authentication implementation that uses server-side sessions
// stored in cookies. Maintained for backward compatibility but deprecated in favor
// of JWT-based authentication (AuthMiddleware).
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.HandlerFunc: Middleware function that validates session authentication
//
// Authentication Method:
//   - Uses Gorilla sessions with cookie storage
//   - Server-side session state management
//   - Cookie-based authentication tokens
//
// Session Validation:
//  1. Retrieves session from cookie store using server secret
//  2. Checks "authenticated" boolean flag in session
//  3. Extracts user object from session if available
//  4. Injects user context for downstream handlers
//
// Context Values Added:
//   - "user": user.User object from session (if available)
//
// Deprecation Notice:
//   - This middleware is deprecated in favor of JWT-based AuthMiddleware
//   - Maintained for backward compatibility with legacy clients
//   - New implementations should use AuthMiddleware for stateless authentication
//
// Security Limitations:
//   - Requires server-side session storage
//   - Not suitable for distributed/stateless architectures
//   - Cookie-based authentication has CSRF vulnerabilities
//   - Less secure than JWT-based authentication
func AuthMiddlewareV1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Retrieve server secret key for session validation
		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			return // Cannot validate sessions without secret key
		}

		// Step 2: Initialize cookie store with server secret
		var store = sessions.NewCookieStore([]byte(SESSION_KEY))

		// Step 3: Retrieve session from cookie store
		session, err := store.Get(r, "session-auth")
		if err != nil {
			return // Invalid or corrupted session
		}

		// Step 4: Check authentication status from session
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			return // User not authenticated or invalid session
		}

		// Step 5: Extract user object from session if available
		userID, ok := session.Values["user"].(user.User)
		if ok {
			ctx := context.WithValue(r.Context(), "user", userID)
			r = r.WithContext(ctx)
		}

		// Step 6: Continue to next handler with session-based authentication
		next.ServeHTTP(w, r)
	}
}

// ErrorHandlerMiddleware is the central error handling point for all API handlers.
// It catches errors returned from handlers, logs them, and sends appropriate HTTP responses.
// This ensures consistent error logging and response formatting across the entire API.
//
// Error Handling Flow:
//  1. Catches errors returned from handlers (via c.Next())
//  2. Checks if error is a ServerError (has status code)
//  3. Logs error with appropriate level (ERROR/WARN based on status code)
//  4. Sends JSON response with error details
//  5. Returns nil to prevent double handling
//
// Usage:
//   - Applied after LoggerMiddleware to have access to request context
//   - Handlers should return *errors.ServerError or regular errors
//   - Regular errors are wrapped as 500 Internal Server Error
func ErrorHandlerMiddleware(c *fiber.Ctx) error {
	// Execute handlers first
	err := c.Next()

	// If no error, continue normally
	if err == nil {
		return nil
	}

	// Extract context from Fiber locals (set by LoggerMiddleware)
	ctx := context.Background()
	if requestID := c.Locals("request_id"); requestID != nil {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}
	if startTime := c.Locals("start_time"); startTime != nil {
		ctx = context.WithValue(ctx, "start_time", startTime)
	}
	ctx = context.WithValue(ctx, "component", "api")

	// Try to extract ServerError from error chain
	var serverErr *errors.ServerError
	if Errors.As(err, &serverErr) {

		// We have a ServerError with status code
		ctx = context.WithValue(ctx, "status_code", serverErr.StatusCode)

		// Log based on status code severity
		if serverErr.StatusCode >= 500 {
			LogError(ctx, serverErr)
		} else {
			// 4xx errors are client errors, log as WARN
			LogWarn(ctx, serverErr)
		}

		// if c.Response().Header.ContentLength() > 0 || c.Response().StatusCode() != 0 {
		// 	log.Println("Log from ErrorHandlerMiddleware: Response already sent")
		// 	// Response already sent, just log
		// 	if serverErr.StatusCode >= 500 {
		// 		LogError(ctx, serverErr)
		// 	} else {
		// 		LogWarn(ctx, serverErr)
		// 	}
		// 	c.Locals("error_handled", true)
		// 	return nil
		// }

		// Send JSON response
		c.Locals("error_handled", true)
		return c.Status(serverErr.StatusCode).JSON(fiber.Map{
			"error":      serverErr.Message,
			"error_code": serverErr.Code,
		})
	}

	// Try to extract AppError (no status code)
	var appErr *errors.AppError
	if Errors.As(err, &appErr) {
		// Convert AppError to ServerError with 500 status
		serverErr = appErr.ToServerError(fiber.StatusInternalServerError)
		ctx = context.WithValue(ctx, "status_code", serverErr.StatusCode)
		LogError(ctx, serverErr)

		c.Locals("error_handled", true)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":      serverErr.Message,
			"error_code": serverErr.Code,
		})
	}

	// Unknown error type - wrap it
	serverErr = errors.WrapServer(
		err,
		errors.InternalError,
		"Internal server error",
		fiber.StatusInternalServerError,
	)
	ctx = context.WithValue(ctx, "status_code", serverErr.StatusCode)
	LogError(ctx, serverErr)

	c.Locals("error_handled", true)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":      serverErr.Message,
		"error_code": serverErr.Code,
	})
}
