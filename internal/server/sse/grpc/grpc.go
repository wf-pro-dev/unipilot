// Package sse provides gRPC server functionality for the SSE notification system.
// Implements a gRPC service that receives notification requests from the main API server
// and forwards them to the SSE server for real-time delivery to connected clients.
package sse

import (
	"context"
	"net"

	"unipilot/internal/server"
	serverSSE "unipilot/internal/server/sse"
	"unipilot/internal/server/sse/grpc/notifications"

	"google.golang.org/grpc"
)

// StartGRPCServer initializes and starts the gRPC notification service on port 9000.
// Creates a gRPC server that accepts notification requests from the main API server
// and forwards them to the SSE server for real-time delivery to connected clients.
// This enables decoupled communication between the API server and SSE service.
//
// Parameters:
//   - sseServer: SSE server instance for delivering notifications (*serverSSE.SSEServer, required)
//
// Server Configuration:
//   - Port: 9000 (dedicated gRPC service port)
//   - Protocol: gRPC over TCP
//   - Service: notifications.NotificationsService
//   - Methods: SendNotification, Heartbeat
//
// Architecture:
//   - Bridges API server (HTTP) and SSE server (WebSocket-like)
//   - Enables microservice communication via gRPC
//   - Provides reliable notification delivery mechanism
//
// Side Effects:
//   - Starts gRPC server on port 9000 (blocking operation)
//   - Registers notification service with gRPC server
//   - Logs server startup and listens indefinitely
//   - Fatal error if port 9000 is unavailable or server fails
//
// Error Handling:
//   - Network listener errors cause fatal program termination
//   - Port conflicts prevent service startup
//   - Server errors during operation cause fatal termination
//
// Usage:
//   - Called during SSE service initialization
//   - Runs in main goroutine (blocking)
//   - Requires SSE server instance for notification forwarding
func StartGRPCServer(sseServer *serverSSE.SSEServer) {
	// Step 1: Create TCP listener on port 9000 for gRPC connections
	ctx := context.WithValue(context.Background(), "component", "grpc")
	lis, err := net.Listen("tcp4", "0.0.0.0:9000")
	if err != nil {
		server.LogFatal(ctx, "Failed to listen on port 9000", err,
			"tags", []string{"system", "network", "high"},
			"error_type", "network",
		)
		return
	}

	// Step 2: Create notification service instance with SSE server reference
	s := notifications.Server{
		SSE: sseServer, // SSE server for forwarding notifications
	}

	// Step 3: Initialize gRPC server with default configuration
	grpcServer := grpc.NewServer()

	// Step 4: Register notification service with gRPC server
	notifications.RegisterNotificationsServiceServer(grpcServer, &s)

	// Step 5: Log server startup for monitoring and debugging
	server.LogInfo(ctx, "gRPC server starting", "port", 9000,
		"tags", []string{"system", "network", "high"},
	)

	// Step 6: Start serving gRPC requests (blocking operation)
	if err := grpcServer.Serve(lis); err != nil {
		// Fatal error if server fails during operation
		server.LogFatal(ctx, "Failed to serve gRPC on port 9000", err,
			"tags", []string{"system", "network", "high"},
			"error_type", "network",
		)
	}

}
