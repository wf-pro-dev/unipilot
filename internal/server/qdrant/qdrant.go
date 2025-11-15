package qdrant

import (
	"github.com/qdrant/go-client/qdrant"
)

func NewQdrantClient() (*qdrant.Client, error) {

	// The Go client uses Qdrant's gRPC interface
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "192.168.86.23",
		Port: 6334,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
