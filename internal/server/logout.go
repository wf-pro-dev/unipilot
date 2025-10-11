package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	
	"unipilot/internal/secrets"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Logout: %w",err))
		return
	}

	var store = sessions.NewCookieStore([]byte(SESSION_KEY))

	session, _ := store.Get(r, "session-auth")

	// Check if user was actually logged in
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		PrintERROR(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	// Clear session values
	session.Values["authenticated"] = false
	delete(session.Values, "user_id")

	// Optionally, expire the session cookie immediately
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}
