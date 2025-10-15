package grpc

import (
	"context"
	"fmt"
	"log"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/secrets"
	"unipilot/internal/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conn *grpc.ClientConn

func NewGRPCClient() *notifications.NotificationsServiceClient {

	address, err := secrets.GetEnvVar("GRPC_SSE_ADDRESS")
	if err != nil {
		log.Fatalf("did not get port: %v", err)
	}

	conn, err = grpc.Dial(fmt.Sprintf("%s:9000", address), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := notifications.NewNotificationsServiceClient(conn)
	message, err := c.Heartbeat(context.Background(), &notifications.Message{Body: "heartbeat"})
	if err != nil {
		log.Fatalf("did not send heartbeat: %v", err)
	}
	server.PrintLOG([]string{"GRPC"}, fmt.Sprintf("Heartbeat response: %s", message.Body))

	return &c
}

func CloseGRPCClient() {

	if conn == nil {
		return
	}

	if err := conn.Close(); err != nil {
		log.Fatalf("did not close connection: %v", err)
	}
	server.PrintLOG([]string{"GRPC"}, "Closing gRPC connection")
}
