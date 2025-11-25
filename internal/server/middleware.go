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

//--------------------------------------------------
// DBMiddleware adds the database connection to the request context
//--------------------------------------------------

func DBMiddleware(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db)
		next(w, r.WithContext(ctx))
	}
}

//--------------------------------------------------
// LoggerMiddleware adds the logger to the request context
//--------------------------------------------------

func LoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		ctx = context.WithValue(ctx, "start_time", startTime)

		next(w, r.WithContext(ctx))

		userID := r.Context().Value("user_id")
		if userID == nil {
			userID = 0
		}
		userID, ok := userID.(uint)
		if !ok {
			userID = 0
		}

		duration := time.Since(startTime).Milliseconds()
		log_message := fmt.Sprintf("%s %d %s %s %s %dms", requestID, userID, r.Method, r.URL.Path, r.RemoteAddr, duration)

		Logger.Debugf(log_message)

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

// ---------------------------------------------------
// AuthMiddleware checks if the user is authenticated
// ---------------------------------------------------

type Claims struct {
	User user.User `json:"user"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return
		}

		// Get session key
		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(SESSION_KEY), nil // Use same secret as main service
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		user := claims.User

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "user_id", user.ID)

		LoggerMiddleware(next)(w, r.WithContext(ctx))
	}
}

// D
func AuthMiddlewareV1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			return
		}

		var store = sessions.NewCookieStore([]byte(SESSION_KEY))

		session, err := store.Get(r, "session-auth")
		if err != nil {
			return
		}

		// Check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			return
		}

		// You can also add the user ID to the request context if needed
		userID, ok := session.Values["user"].(user.User)
		if ok {
			ctx := context.WithValue(r.Context(), "user", userID)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	}
}
