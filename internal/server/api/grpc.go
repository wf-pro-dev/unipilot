package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"unipilot/internal/errors"
	"unipilot/internal/secrets"
	"unipilot/internal/server"
	"unipilot/internal/server/sse/grpc/messages"
)

var conn *grpc.ClientConn
var GrpcClient *messages.MessageServiceClient

func NewGRPCClient() error {

	address, err := secrets.GetEnvVar("GRPC_SSE_ADDR")
	if err != nil {
		return errors.Wrap(err, errors.ConfigEnvVarNotFound, "cannot get grpc sse address")
	}

	conn, err = grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, errors.GRPCClientFailed, "failed to create gRPC client")
	}

	c := messages.NewMessageServiceClient(conn)
	message, err := c.SendHeartbeat(context.Background(), &messages.Heartbeat{Body: "heartbeat"})
	if err != nil {
		return errors.Wrap(err, errors.GRPCUnreachable, "cannot send heartbeat")
	}
	ctx := context.WithValue(context.Background(), "component", "grpc")
	server.LogDebug(ctx, "Heartbeat response received", "message", message.Body,
		"tags", []string{"system", "network", "low"},
	)

	GrpcClient = &c

	return nil
}

func CloseGRPC() error {

	if conn != nil {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}
