package grpc

import (
	"context"
	"fmt"
	"log"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/secrets"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient() *notifications.NotificationsServiceClient {

	address, err := secrets.GetEnvVar("GRPC_SSE_ADDRESS")
	if err != nil {
		log.Fatalf("did not get port: %v", err)
	}

	var conn *grpc.ClientConn
	conn, err = grpc.Dial(fmt.Sprintf("%s:9000", address), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := notifications.NewNotificationsServiceClient(conn)
	c.Heartbeat(context.Background(), &notifications.Message{Body: "heartbeat"})

	return &c
}
