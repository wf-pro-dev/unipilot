// Package notifications implements the gRPC notification service for real-time messaging.
// Provides the server-side implementation of the NotificationsService gRPC interface,
// handling notification requests from the API server and forwarding them to the SSE server
// for delivery to connected clients.
package notifications

import (
	"context"
	"log"

	"unipilot/internal/models"
	"unipilot/internal/models/notifications"
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
	UnimplementedNotificationsServiceServer
	// SSE server instance for delivering notifications to connected clients
	SSE *sse.SSEServer
}

// SendNotification processes gRPC notification requests and forwards them to the SSE server.
// Converts gRPC notification format to internal notification structure and delivers
// to the target user via their SSE connection if they are connected.
//
// gRPC Method: SendNotification
// Service: notifications.NotificationsService
//
// Parameters:
//   - ctx: Request context for cancellation and tracing (context.Context, required)
//   - notification: gRPC notification message with all delivery details (*Notification, required)
//
// Notification Fields (from gRPC message):
//   - UserId: Target user ID for notification delivery (uint32, required)
//   - SenderId: User who triggered the notification (uint32, required)
//   - Entity: Entity type (assignment, document, course, etc.) (string, required)
//   - EntityId: Specific entity identifier (uint32, required)
//   - Type: Notification type (create, update, delete, etc.) (string, required)
//   - Title: Human-readable notification title (string, required)
//   - Message: Detailed notification message (string, required)
//   - Action: Client action type for UI handling (string, required)
//   - Data: JSON payload with additional context (string, optional)
//
// Returns:
//   - *NotificationResponse: gRPC response with success status
//   - error: nil if successful, error if notification processing fails
//
// Processing Flow:
//  1. Log incoming notification request for debugging
//  2. Convert gRPC message types to internal types
//  3. Forward to SSE server for delivery
//  4. Log delivery attempt with structured data
//  5. Return success status based on delivery result
//
// Error Handling:
//   - JSON marshalling errors from SSE server
//   - Type conversion errors (handled gracefully)
//   - Delivery failures for disconnected users (not considered errors)
//
// Side Effects:
//   - Logs notification request and delivery attempt
//   - Forwards notification to SSE server for client delivery
//   - Updates monitoring metrics for notification processing
func (s *Server) SendNotification(ctx context.Context, notification *Notification) (*NotificationResponse, error) {
	// Step 1: Log incoming notification request for debugging and monitoring
	log.Printf("[Notifications] Sending notification: %v", notification)

	// Step 2: Convert gRPC types to internal types and forward to SSE server
	err := s.SSE.SendNotification(
		uint(notification.UserId),                         // Convert uint32 to uint
		uint(notification.SenderId),                       // Convert uint32 to uint
		models.Entity(notification.Entity),                // Convert string to Entity enum
		uint(notification.EntityId),                       // Convert uint32 to uint
		notifications.NotificationType(notification.Type), // Convert string to NotificationType enum
		notification.Title,                                // Pass string directly
		notification.Message,                              // Pass string directly
		notification.Action,                               // Pass string directly
		notification.Data,                                 // Pass JSON string directly
	)

	// Step 3: Log notification delivery attempt with structured data for monitoring
	server.LogDebug(ctx, "Notification sent: ",
		"title", notification.Title,
		"sender_id", notification.SenderId,
		"user_id", notification.UserId,
		"tags", []string{"SSE", "GRPC", "NOTIFICATION"},
	)

	// Step 4: Return gRPC response with success status based on delivery result
	return &NotificationResponse{Success: err == nil}, err
}

// Heartbeat provides a health check endpoint for the gRPC notification service.
// Allows clients to verify service availability and measure round-trip latency.
// Used for service discovery, load balancing, and monitoring system health.
//
// gRPC Method: Heartbeat
// Service: notifications.NotificationsService
//
// Parameters:
//   - ctx: Request context for cancellation and tracing (context.Context, required)
//   - message: Heartbeat message from client (*Message, required)
//
// Message Fields:
//   - Body: Heartbeat message content (string, optional)
//
// Returns:
//   - *Message: Response message confirming heartbeat received
//   - error: Always nil (heartbeat never fails)
//
// Response:
//   - Body: "heartbeat received" (constant confirmation message)
//
// Use Cases:
//   - Service health monitoring
//   - Load balancer health checks
//   - Network connectivity verification
//   - Round-trip latency measurement
//
// Side Effects:
//   - Logs heartbeat request for monitoring
//   - Updates service availability metrics
//   - No persistent state changes
func (s *Server) Heartbeat(ctx context.Context, message *Message) (*Message, error) {
	// Step 1: Log heartbeat request for service monitoring and debugging
	server.LogDebug(ctx, "Heartbeat from client: ",
		"body", message.Body,
		"tags", []string{"GRPC", "HEARTBEAT"},
	)

	// Step 2: Return confirmation message to client
	return &Message{Body: "heartbeat received"}, nil
}
