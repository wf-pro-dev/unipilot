package grpc

import (
	"context"
	"log"
	"unipilot/internal/server/sse/grpc/notifications"

	"unipilot/internal/secrets"
	"unipilot/internal/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var conn *grpc.ClientConn

func NewGRPCClient() *notifications.NotificationsServiceClient {

	address, err := secrets.GetEnvVar("GRPC_SSE_ADDR")
	if err != nil {
		log.Fatalf("did not get port: %v", err)
	}

	conn, err = grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := notifications.NewNotificationsServiceClient(conn)
	message, err := c.Heartbeat(context.Background(), &notifications.Message{Body: "heartbeat"})
	if err != nil {
		log.Fatalf("did not send heartbeat: %v", err)
	}
	server.LogDebug(context.Background(), "Heartbeat response: ",
		"response", message.Body,
		"tags", []string{"GRPC", "HEARTBEAT"},
	)

	return &c
}

func CloseGRPCClient() {

	ctx := context.Background()
	if conn != nil {
		if err := conn.Close(); err != nil {
			server.LogError(ctx, "Failed to close gRPC connection", err,
				"tags", []string{"GRPC", "CLOSE"},
			)
		}
	}
	server.LogDebug(ctx, "Closed gRPC connection",
		"tags", []string{"GRPC", "CLOSE"},
	)
}
