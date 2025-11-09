package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"unipilot/internal/models/user"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
)

func HandleRefreshToken(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("user").(user.User)

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error getting session key: %s", err.Error()))
		return
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error creating new token: %s", err.Error()))
		return
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, server.Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	}).SignedString([]byte(SESSION_KEY))
	if err != nil {
		server.PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("error creating new refresh token: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	server.PrintLOG([]string{"SUCCESS", "TOKEN", "REFRESH"}, fmt.Sprintf("Token refreshed successfully for user id: %d", user.ID))
}
