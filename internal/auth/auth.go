package auth

import (
	"net/http"
	"unipilot/internal/sse"
	"unipilot/internal/storage"

	"gorm.io/gorm"
)

type Auth struct {
	Client  *http.Client
	SSE     *sse.SSE
	LocalDB *gorm.DB
}

func NewAuth() *Auth {

	return &Auth{}
}

func (a *Auth) IsAuthenticated() bool {
	creds, err := storage.GetCurrentUser()
	if err != nil {
		return false
	}
	return creds != nil
}
