package server

import (
	"encoding/json"
	"net/http"
	"unipilot/internal/server"
)

// LogoutHandler handles user logout requests.
// This is a stateless logout implementation that simply acknowledges the logout request.
// Since the system uses JWT tokens (which are stateless), actual token invalidation
// must be handled client-side by discarding the tokens.
//
// HTTP Method: POST
// Content-Type: Not required (no request body expected)
//
// Request Body: None required
//
// Response (200 OK):
//   - message: Success confirmation message
//
// Security Notes:
//   - This is a client-side logout implementation
//   - JWT tokens remain valid until expiration (15 min for access, 30 days for refresh)
//   - For true server-side logout, consider implementing a token blacklist
//   - Client should discard tokens after receiving this response
//
// Side Effects:
//   - Logs logout event for audit trail
//   - No server-side session state changes (stateless design)
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	// Step 1: Enforce POST-only endpoint for security (registration should never be GET)
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Step 2: Send successful logout acknowledgment (no server-side token invalidation needed)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})

	// Step 3: Log logout event for security audit trail and monitoring
	server.LogInfo(r.Context(), "Logout successful",
		"tags", []string{"LOGOUT"},
	)
}
