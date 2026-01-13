package main

import (
	"log"

	"unipilot/internal/errors"
	server "unipilot/internal/server/api"
)

func main() {
	err := server.StartServer()
	if err != nil {
		// Check if error is already wrapped with error system
		if errors.HasAppError(err) {
			log.Fatalf("Failed to start api: %v", err)
		} else {
			// Wrap unknown errors with error system
			wrappedErr := errors.Wrap(err, errors.APIStartFailed, "Failed to start API server")
			log.Fatalf("Failed to start api: %v", wrappedErr)
		}
	}
}
