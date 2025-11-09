package main

import (
	server "unipilot/internal/server/sse"
	grpc "unipilot/internal/server/sse/grpc"
)

func main() {
	sseServer := server.StartSSEServer()
	grpc.StartGRPCServer(sseServer)
}
