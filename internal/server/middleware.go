package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"

	"unipilot/internal/secrets"

	"unipilot/internal/models/user"
)

// DBMiddleware injects the database connection into the HTTP request context.
// Provides database access to all downstream handlers without requiring explicit
// parameter passing. Essential middleware for all database-dependent endpoints.
//
// Parameters:
//   - db: GORM database connection instance
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.HandlerFunc: Middleware function that adds database to request context
//
// Context Values Added:
//   - "db": *gorm.DB instance for database operations
//
// Usage:
//   - Applied to all API routes that require database access
//   - Handlers can retrieve database via: r.Context().Value("db").(*gorm.DB)
//
// Security Considerations:
//   - Database connection is shared across requests (connection pooling)
//   - No user-specific database isolation (handled at application level)
func DBMiddleware(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Inject database connection into request context for downstream handlers
		ctx := context.WithValue(r.Context(), "db", db)
		next(w, r.WithContext(ctx))
	}
}

// LoggerMiddleware provides comprehensive request logging and performance tracking.
// Generates unique request IDs, tracks request duration, and logs detailed request
// information for debugging, monitoring, and audit purposes.
//
// Parameters:
//   - next: The next HTTP handler in the middleware chain
//
// Returns:
//   - http.HandlerFunc: Middleware function that adds logging context and tracks requests
//
// Context Values Added:
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
func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate unique request ID and capture start time for tracking
		startTime := time.Now()
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		ctx = context.WithValue(ctx, "start_time", startTime)

		// Execute the next handler with enriched context
		next(w, r.WithContext(ctx))

		// Extract user ID from context (if available from authentication)
		userID := r.Context().Value("user_id")
		if userID == nil {
			userID = 0
		}
		userID, ok := userID.(uint)
		if !ok {
			userID = 0
		}

		// Calculate request duration and format comprehensive log message
		duration := time.Since(startTime).Milliseconds()
		log_message := fmt.Sprintf("%s %d %s %s %s %dms", requestID, userID, r.Method, r.URL.Path, r.RemoteAddr, duration)

		// Log to console for development debugging
		Logger.Debugf(log_message)

		// Log structured data to file for production monitoring and analysis
		FileLogger.Infow(log_message,
			"request_id", requestID,
			"user_id", userID,
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration", duration,
		)

	}
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

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Extract JWT token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return // No auth header - allow public endpoints to handle
		}

		// Step 2: Parse Bearer token format ("Bearer <token>")
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return // Invalid format - not a Bearer token
		}

		// Step 3: Retrieve server secret key for token validation
		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			return // Cannot validate without secret key
		}

		// Step 4: Parse and validate JWT token with server secret
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(SESSION_KEY), nil // Use same secret as token generation
		})

		// Step 5: Validate token parsing and signature
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Step 6: Extract and validate claims structure
		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Step 7: Extract user object from validated claims
		user := claims.User

		// Step 8: Inject user context into request for downstream handlers
		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "user_id", user.ID)

		// Step 9: Continue to next handler with authenticated user context and logging
		LoggerMiddleware(next)(w, r.WithContext(ctx))
	}
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
