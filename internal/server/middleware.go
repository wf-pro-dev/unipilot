package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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
			PrintERROR(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			PrintERROR(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		// Get session key
		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Register: %v", err))
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

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user", claims.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// D
func AuthMiddlewareV1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Register: %v", err))
			return
		}

		var store = sessions.NewCookieStore([]byte(SESSION_KEY))

		session, err := store.Get(r, "session-auth")
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %v", err))
			return
		}

		// Check if user is authenticated
		auth, ok := session.Values["authenticated"].(bool)
		if !ok || !auth {
			PrintERROR(w, http.StatusUnauthorized, "Unauthorized - please login")
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
