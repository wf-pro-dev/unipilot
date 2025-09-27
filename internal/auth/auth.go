package auth

import (
	"log"
	"net/http"
	"unipilot/internal/client"
	"unipilot/internal/models/user"
	"unipilot/internal/services/utils"
	"unipilot/internal/sse"
)

type Auth struct {
	Client *http.Client
	SSE    *sse.SSE
	User   *user.User
}

func NewAuth() *Auth {
	newAuth := &Auth{
		Client: nil,
		SSE:    nil,
		User:   nil,
	}

	currentUser, err := utils.GetUserFromFile()
	if err == nil {
		newAuth.User = currentUser

		newAuth.Client, err = client.NewClientWithCookies()
		if err != nil {
			log.Println("Error creating client: ", err)
		}

		log.Printf("Auth: %v", newAuth.Client)
	}

	return newAuth
}

func (a *Auth) IsAuthenticated() bool {
	return a.User != nil
}
