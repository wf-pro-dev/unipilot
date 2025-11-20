package qdrant

import (
	"unipilot/internal/secrets"

	"github.com/qdrant/go-client/qdrant"
)

func NewQdrantClient() (*qdrant.Client, error) {

	qdrantHost, err := secrets.GetEnvVar("QDRANT_HOST")
	if err != nil {
		return nil, err
	}

	// The Go client uses Qdrant's gRPC interface
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: qdrantHost,
		Port: 6334,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
