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

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(user.User)

	SESSION_KEY, err := secrets.GetEnvVar("SESSION_KEY")
	if err != nil {
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error getting session key",
			"tags", []string{"TOKEN", "SESSION_KEY"},
		)
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
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating access token",
			"tags", []string{"TOKEN", "ACCESS_TOKEN"},
		)
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
		server.ResponseError(r.Context(),
			w, err, http.StatusInternalServerError, "Error creating refresh token",
			"tags", []string{"TOKEN", "REFRESH_TOKEN"},
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
		"token":         accessToken,
		"refresh_token": refreshToken,
	})

	server.LogInfo(r.Context(),
		"Token refreshed successfully",
		"tags", []string{"TOKEN"},
	)
}
