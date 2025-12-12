package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

// RefreshTokenHandler handles JWT token refresh requests.
// Generates new access and refresh tokens for authenticated users to maintain
// continuous session access without requiring re-authentication. Uses the existing
// user context from the refresh token to create new tokens with fresh expiration times.
//
// HTTP Method: POST
// Content-Type: Not required (no request body expected)
//
// Request Body: None required (user context extracted from existing refresh token)
//
// Response (200 OK):
//   - message: Success message
//   - token: New JWT access token (expires in 15 minutes)
//   - refresh_token: New JWT refresh token (expires in 30 days)
//
// Authentication: Required (AuthMiddleware) - validates existing refresh token
//
// Security Features:
//   - Generates completely new tokens (not just extends expiration)
//   - Maintains same user context but with fresh timestamps
//   - Uses secure JWT signing with HS256 algorithm
//   - Appropriate token lifespans (15 min access, 30 days refresh)
//
// Error Responses:
//   - 401 Unauthorized: Invalid or expired refresh token
//   - 500 Internal Server Error: Session key retrieval or token generation failure
//
// Side Effects:
//   - Creates new JWT tokens with current timestamps
//   - Logs token refresh events for audit trail
//   - Previous tokens remain valid until natural expiration
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Enforce POST-only endpoint for security (refresh should never be GET)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Extract user context from request context (validated by AuthMiddleware)
	user := r.Context().Value("user").(user.User)

	// Step 3: Retrieve JWT signing key from environment for secure token generation
	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error getting session key",
			"tags", []string{"TOKEN", "SESSION_KEY"},
		)
		return
	}

	// Step 4: Generate new access token with 15-minute expiration for API access
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating access token",
			"tags", []string{"TOKEN", "ACCESS_TOKEN"},
		)
		return
	}

	// Step 5: Generate new refresh token with 30-day expiration for long-term sessions
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating refresh token",
			"tags", []string{"TOKEN", "REFRESH_TOKEN"},
		)
		return
	}

	// Step 6: Send successful response with both new tokens
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	// Step 7: Log successful token refresh for audit trail and monitoring
	server.LogInfo(r.Context(), "Token refreshed", "user_id", user.ID,
		"tags", []string{"auth", "auth", "medium"})
}
