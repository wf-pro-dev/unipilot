// Package qdrant provides client functionality for connecting to Qdrant vector database.
// Qdrant is used for vector similarity search and storage in the unipilot system.
package qdrant

import (
	"unipilot/internal/secrets"

	"github.com/qdrant/go-client/qdrant"
)

// NewQdrantClient creates and returns a new Qdrant client instance.
// The client connects to the Qdrant vector database using gRPC protocol on port 6334.
//
// The function retrieves the Qdrant host from environment variables and establishes
// a connection using the official Qdrant Go client library.
//
// Returns:
//   - *qdrant.Client: A configured Qdrant client ready for vector operations
//   - error: An error if the QDRANT_HOST environment variable is missing or
//     if the client connection fails
//
// Environment Variables Required:
//   - QDRANT_HOST: The hostname or IP address of the Qdrant server
//
// Example:
//
//	client, err := NewQdrantClient()
//	if err != nil {
//	    log.Fatal("Failed to create Qdrant client:", err)
//	}
//	defer client.Close()
func NewQdrantClient() (*qdrant.Client, error) {

	// Step 1: Retrieve Qdrant server configuration from environment
	// This allows for different hosts in dev/staging/production environments
	qdrantHost, err := secrets.GetEnvVar("QDRANT_HOST")
	if err != nil {
		return nil, err
	}

	// Step 2: Establish gRPC connection to Qdrant vector database
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: qdrantHost,
		Port: 6334, // Standard Qdrant gRPC port for vector operations
	})
	if err != nil {
		// Connection failure could be due to network issues or wrong host/port
		return nil, err
	}

	// Step 3: Return ready-to-use client for vector operations
	return client, nil
}
