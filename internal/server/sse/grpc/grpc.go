package sse

import (
	"log"
	"net"

	serverSSE "unipilot/internal/server/sse"
	"unipilot/internal/server/sse/grpc/notifications"

	"google.golang.org/grpc"
)

func StartGRPCServer(sseServer *serverSSE.SSEServer) {
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen on port 9000: %v", err)
	}

	s := notifications.Server{
		SSE: sseServer,
	}

	grpcServer := grpc.NewServer()
	notifications.RegisterNotificationsServiceServer(grpcServer, &s)
	log.Println("GRPC server listening on :9000...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve on port 9000: %v", err)
	}

}
