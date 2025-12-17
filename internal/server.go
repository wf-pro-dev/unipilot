package main

import (
	"log"

	server "unipilot/internal/server/api"
)

func main() {
	err := server.StartServer()
	if err != nil {
		log.Fatalf("Failed to start api: %v", err)
	}
}
