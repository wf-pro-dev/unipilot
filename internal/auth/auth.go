package auth

import (
	"net/http"
	"unipilot/internal/client"
	"unipilot/internal/sse"

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

// IsAuthenticated checks if the user is currently authenticated
func (a *Auth) IsAuthenticated() bool {
	// If we don't have a client, try to initialize it with saved cookies
	if a.Client == nil {
		httpClient, err := client.NewClient()
		if err != nil {
			return false
		}
		a.Client = httpClient
	}

	// Make a request to the user endpoint to check authentication
	resp, err := a.Client.Get("https://newsroom.dedyn.io/acc-homework/user")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If we get a 200 OK, we're authenticated
	// If we get a 401 Unauthorized, we're not authenticated
	return resp.StatusCode == http.StatusOK
}
