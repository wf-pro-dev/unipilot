package main

import (
	"unipilot/internal/server"
	"unipilot/internal/server/sse"
	grpc "unipilot/internal/server/sse/grpc"
)

func main() {
	server.InitLogger()
	sseServer := sse.StartSSEServer()
	grpc.StartGRPCServer(sseServer)

}
