// Package notifications implements the gRPC notification service for real-time messaging.
package messages

import (
	"context"

	"unipilot/internal/models"
	"unipilot/internal/server"
	sse "unipilot/internal/server/sse"
)

// Server implements the gRPC NotificationsService interface for handling notification requests.
// Acts as a bridge between the API server (which sends notifications via gRPC) and the
// SSE server (which delivers notifications to connected clients via WebSocket-like connections).
//
// Fields:
//   - UnimplementedNotificationsServiceServer: gRPC generated base implementation
//   - SSE: Reference to SSE server for notification delivery
//
// Architecture Role:
//   - Receives gRPC notification requests from API server
//   - Converts gRPC message format to SSE notification format
//   - Forwards notifications to SSE server for client delivery
//   - Provides service health monitoring via heartbeat endpoint
type Server struct {
	// UnimplementedNotificationsServiceServer provides forward compatibility for gRPC
	UnimplementedMessageServiceServer
	// SSE server instance for delivering notifications to connected clients
	SSE *sse.SSEServer
}

func (s *Server) SendMessage(ctx context.Context, message *Message) (*Response, error) {

	// Step 2: Convert gRPC types to internal types and forward to SSE server
	err := s.SSE.SendMessage(
		uint(message.SenderId),           // Convert uint32 to uint
		uint(message.ReceiverId),         // Convert uint32 to uint
		message.Title,                    // Pass string directly
		message.Message,                  // Pass string directly
		message.Data,                     // Pass JSON string directly
		models.MessageType(message.Type), // Convert string to MessageType enum
	)

	// Step 3: Log message delivery attempt with structured data for monitoring
	ctx = context.WithValue(ctx, "component", "grpc")
	server.LogDebug(ctx, "sender_id", message.SenderId, "receiver_id", message.ReceiverId, "message", message.Title)

	// Step 4: Return gRPC response with success status based on delivery result
	return &Response{Success: err == nil}, err
}

// Heartbeat provides a health check endpoint for the gRPC notification service.
func (s *Server) SendHeartbeat(ctx context.Context, heartbeat *Heartbeat) (*Heartbeat, error) {
	// Step 1: Log heartbeat request for service monitoring and debugging
	ctx = context.WithValue(ctx, "component", "grpc")
	server.LogDebug(ctx, "Heartbeat received",
		"tags", []string{"system", "network", "low"},
	)

	// Step 2: Return confirmation message to client
	return &Heartbeat{Body: "heartbeat received"}, nil
}
