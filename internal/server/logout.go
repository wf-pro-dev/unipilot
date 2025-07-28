package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/viper"

	"github.com/gorilla/sessions"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error reading config file: %w", err))
		return
	}

	SESSION_KEY := viper.GetString("SESSION_KEY")

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
