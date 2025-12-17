// Package sse provides Server-Sent Events functionality for real-time notifications.
// instant updates to connected clients when course-related activities occur.
package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/notifications"
	"unipilot/internal/server"
)

// SSEClient represents a connected client receiving Server-Sent Events.
// Each client maintains a persistent HTTP connection and receives real-time
// notifications through a buffered message channel.
type SSEClient struct {
	// UserID uniquely identifies the authenticated user for this connection
	UserID uint
	// Messages is a buffered channel for queuing outbound notifications
	Messages chan []byte
	// Connected indicates if the client connection is active
	Connected bool
	// LastActive tracks the last heartbeat or message sent to detect stale connections
	LastActive time.Time
}

// SSEServer manages multiple SSE client connections and handles message broadcasting.
// It provides thread-safe operations for adding/removing clients and sending notifications
// to specific users or all connected clients.
type SSEServer struct {
	// clients maps user IDs to their active SSE connections
	clients map[uint]*SSEClient
	// mu provides thread-safe access to the clients map
	mu sync.RWMutex
}

// NewSSEServer creates and initializes a new SSE server instance.
// Returns a server ready to accept client connections and manage notifications.
func NewSSEServer() *SSEServer {
	return &SSEServer{
		clients: make(map[uint]*SSEClient),
	}
}

// HealthHandler provides a health check endpoint for the SSE server.
// Returns JSON response with current server status and timestamp for monitoring.
// Used by load balancers and monitoring systems to verify service availability.
//
// HTTP Method: GET
// Content-Type: application/json (response)
//
// Request Body: None required
//
// Response (200 OK):
//   - status: "healthy" (string, constant)
//   - timestamp: Current server time in RFC3339 format (string)
//
// Authentication: Not required (public health check endpoint)
//
// Security Features:
//   - No sensitive information exposed in response
//   - Safe for public access without authentication
//   - Minimal resource usage for high-frequency monitoring
//
// Use Cases:
//   - Load balancer health checks
//   - Monitoring system service verification
//   - Docker container health status
//   - Kubernetes readiness/liveness probes
func HealthHandler(c *fiber.Ctx) error {
	// Step 3: Write health status JSON with current timestamp
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// StartSSEServer initializes and starts the SSE HTTP server on port 3000.
// Sets up health check and authenticated SSE endpoints, then starts listening
// for connections in a separate goroutine to avoid blocking the caller.
// Returns the server instance for external notification sending.
//
// Server Configuration:
//   - Port: 3000 (dedicated SSE service port)
//   - Endpoints: /health (public), /unipilot/sse/v1 (authenticated)
//   - Middleware: JWT authentication required for SSE connections
//   - Concurrency: Non-blocking startup using goroutine
//
// Returns:
//   - *SSEServer: Initialized server instance for sending notifications
//
// Side Effects:
//   - Starts HTTP server on port 3000 in background goroutine
//   - Registers HTTP handlers for health check and SSE endpoints
//   - Logs server startup message for monitoring
//   - Fatal error if port 3000 is unavailable
//
// Error Handling:
//   - Server startup errors cause fatal program termination
//   - Port conflicts will prevent service startup
//   - Authentication middleware errors handled per-request
func StartSSEServer() *SSEServer {
	// Step 1: Create new SSE server instance with empty client map
	sseServer := NewSSEServer()

	// Step 2: Initialize Fiber app for SSE server
	app := fiber.New()

	// Step 3: Register public health check endpoint for monitoring
	app.Get("/health", HealthHandler)

	// Step 4: Register active clients count endpoint for monitoring
	app.Get("/unipilot/sse/v1/count", sseServer.CountHandler)

	// Step 4: Register authenticated SSE endpoint with JWT middleware
	app.Get("/unipilot/sse/v1", server.AuthMiddleware, sseServer.SSEHandler)

	// Step 6: Start Fiber server in background goroutine to avoid blocking
	go func() {
		if err := app.Listen(":3002"); err != nil {
			// Fatal error if server cannot start (port conflict, permissions, etc.)
			log.Fatalf("SSE server error: %v", err)
		}
	}()

	// Step 7: Return server instance for external notification sending
	return sseServer
}

// AddClient registers a new SSE client connection for the specified user.
// Creates a buffered message channel (100 messages) and adds the client to the server's
// connection pool. If a client already exists for this user, it will be replaced.
// Returns the newly created client instance.
//
// Parameters:
//   - userID: Authenticated user ID from JWT token (uint, required)
//
// Returns:
//   - *SSEClient: Newly created client instance with message channel
//
// Client Configuration:
//   - Message Buffer: 100 messages capacity for burst handling
//   - Connection Status: Initially set to true (connected)
//   - User Isolation: One connection per user (replaces existing)
//
// Thread Safety:
//   - Uses write lock (s.mu.Lock) for exclusive client map access
//   - Safe for concurrent client registration from multiple goroutines
//
// Side Effects:
//   - Adds client to server's connection pool (s.clients map)
//   - Logs new client registration for monitoring
//   - Replaces existing client connection if user already connected
//
// Resource Management:
//   - Creates new buffered channel (100 message capacity)
//   - Previous client channels automatically garbage collected
func (s *SSEServer) AddClient(userID uint) *SSEClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close old client's channel if it exists (prevents goroutine leak)
	if oldClient, exists := s.clients[userID]; exists {
		close(oldClient.Messages)
	}

	// Create new client...
	client := &SSEClient{
		UserID:    userID,
		Messages:  make(chan []byte, 100),
		Connected: true,
	}
	ctx := context.WithValue(context.Background(), "component", "sse")
	server.LogDebug(ctx, "SSE client added", "user_id", userID,
		"tags", []string{"notification", "network", "low"},
	)

	s.clients[userID] = client
	return client
}

// RemoveClient safely disconnects and removes a client from the server.
// Closes the client's message channel to prevent goroutine leaks and removes
// the client from the active connections map. Safe to call multiple times.
//
// Parameters:
//   - userID: User ID of client to remove (uint, required)
//
// Thread Safety:
//   - Uses write lock (s.mu.Lock) for exclusive client map modification
//   - Safe for concurrent client removal from multiple goroutines
//
// Resource Cleanup:
//   - Closes client message channel to signal goroutine termination
//   - Removes client from server's connection pool
//   - Prevents memory leaks from abandoned channels
//
// Side Effects:
//   - Client's message channel closed (triggers goroutine cleanup)
//   - Client removed from active connections map
//   - No effect if client doesn't exist (safe multiple calls)
//
// Error Handling:
//   - Graceful handling of non-existent clients
//   - Channel close is safe even if already closed
func (s *SSEServer) RemoveClient(userID uint) {
	// Step 1: Acquire write lock for exclusive client map modification
	s.mu.Lock()
	defer s.mu.Unlock()

	// Step 2: Check if client exists and perform cleanup
	if client, ok := s.clients[userID]; ok {
		// Step 3: Close message channel to signal goroutine termination
		close(client.Messages)

		// Step 4: Remove client from active connections map
		delete(s.clients, userID)
		ctx := context.WithValue(context.Background(), "component", "sse")
		server.LogDebug(ctx, "SSE client removed", "user_id", userID,
			"tags", []string{"notification", "network", "low"},
		)
	}
}

// SendToUser attempts to send a message to a specific connected user.
// Uses non-blocking channel send to avoid hanging if the client's message buffer
// is full. Logs active client count for monitoring purposes.
// Returns true if message was successfully queued, false if user not connected
// or their message buffer is full.
//
// Parameters:
//   - userID: Target user ID for message delivery (uint, required)
//   - message: JSON-encoded notification message ([]byte, required)
//
// Returns:
//   - bool: true if message queued successfully, false if delivery failed
//
// Delivery Behavior:
//   - Non-blocking send prevents server hanging on slow clients
//   - Message dropped if client buffer full (100 message capacity)
//   - Returns false for disconnected users (graceful degradation)
//
// Thread Safety:
//   - Uses read lock (s.mu.RLock) for concurrent client map access
//   - Safe for concurrent message sending from multiple goroutines
//
// Monitoring:
//   - Logs active client count for connection health tracking
//   - Performance metrics available for message delivery rates
//
// Error Handling:
//   - Graceful failure for disconnected users
//   - Buffer overflow handled by dropping messages (no blocking)
func (s *SSEServer) SendToUser(userID uint, message []byte) bool {
	// Step 1: Acquire read lock for concurrent client map access
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Step 3: Check if target user is connected
	if client, ok := s.clients[userID]; ok {
		// Step 4: Attempt non-blocking message send
		select {
		case client.Messages <- message:
			// Step 5: Message successfully queued for delivery
			return true
		default:
			// Step 6: Channel full, client might be slow - drop message
			return false
		}
	}
	// Step 7: User not connected - return failure
	return false
}

// Broadcast sends a message to all currently connected clients.
// Uses non-blocking sends to prevent slow clients from affecting others.
// Messages are silently dropped for clients with full message buffers.
// Useful for system-wide announcements or maintenance notifications.
//
// Parameters:
//   - message: JSON-encoded message to broadcast ([]byte, required)
//
// Broadcast Behavior:
//   - Sends to all connected clients simultaneously
//   - Non-blocking sends prevent slow clients from affecting others
//   - Messages dropped for clients with full buffers (no blocking)
//
// Thread Safety:
//   - Uses read lock (s.mu.RLock) for concurrent client map access
//   - Safe for concurrent broadcasting from multiple goroutines
//
// Use Cases:
//   - System-wide maintenance announcements
//   - Global notification broadcasts
//   - Emergency alerts to all users
//   - Server shutdown notifications
//
// Performance:
//   - Scales with number of connected clients
//   - No blocking on individual client delivery failures
func (s *SSEServer) Broadcast(message []byte) {
	// Step 1: Acquire read lock for concurrent client map access
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Step 2: Iterate through all connected clients
	for _, client := range s.clients {
		// Step 3: Attempt non-blocking message send to each client
		select {
		case client.Messages <- message:
			// Message successfully queued for this client
		default:
			// Skip if channel is full - don't block on slow clients
		}
	}
}

// logActiveClients outputs the current number of connected clients for monitoring.
// Called during message sending operations to track connection health and usage patterns.
// Uses structured logging for integration with monitoring and alerting systems.
func (s *SSEServer) CountHandler(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return c.JSON(fiber.Map{
		"count":     len(s.clients),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// SSEHandler establishes and manages Server-Sent Event connections for authenticated users.
// Creates a persistent HTTP connection that streams real-time notifications to the client.
// Handles connection lifecycle, heartbeat management, and graceful disconnection.
// Requires authentication middleware to extract user context from JWT tokens.
//
// HTTP Method: GET (long-lived connection)
// Content-Type: text/event-stream (SSE protocol)
//
// Request Requirements:
//   - Authorization: Bearer <JWT_TOKEN> (required)
//   - Accept: text/event-stream (recommended)
//
// Response Headers:
//   - Content-Type: text/event-stream
//   - Cache-Control: no-cache
//   - Connection: keep-alive
//   - Access-Control-Allow-Origin: * (CORS enabled)
//   - X-Accel-Buffering: no (disable proxy buffering)
//
// Authentication: Required (AuthMiddleware) - extracts user from JWT token
//
// Connection Lifecycle:
//  1. JWT token validation and user extraction
//  2. SSE headers setup and flusher initialization
//  3. Client registration and connection establishment
//  4. Message streaming with heartbeat management
//  5. Graceful cleanup on disconnection
//
// Heartbeat System:
//   - 15-second intervals to maintain connection
//   - Prevents proxy timeouts and connection drops
//   - Updates client LastActive timestamp
//
// Message Format:
//   - SSE protocol: "data: {json}\n\n"
//   - JSON-encoded notification payloads
//   - Heartbeat: ": heartbeat\n\n"
//
// Error Handling:
//   - Invalid user context: Silent return (401 handled by middleware)
//   - No flusher support: Silent return (incompatible client)
//   - Client disconnection: Graceful cleanup and logging
//
// Side Effects:
//   - Registers client in server connection pool
//   - Starts heartbeat ticker for connection maintenance
//   - Automatic cleanup on connection termination
func (s *SSEServer) SSEHandler(c *fiber.Ctx) error {
	// Step 1: Extract authenticated user ID from Fiber locals
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		// User context missing - authentication middleware failed
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	// Step 2: Set Server-Sent Events protocol headers
	c.Set("Content-Type", "text/event-stream") // SSE MIME type
	c.Set("Cache-Control", "no-cache")         // Prevent caching
	c.Set("Connection", "keep-alive")          // Maintain connection
	c.Set("Access-Control-Allow-Origin", "*")  // CORS support
	c.Set("X-Accel-Buffering", "no")           // Disable proxy buffering

	// Step 3: Register client connection and setup cleanup
	client := s.AddClient(userID) // Cleanup ticker on exit

	// Step 6: Capture context before entering SetBodyStreamWriter
	// The context may not be accessible inside the callback
	ctxDone := c.Context().Done()

	// Step 7: Use Fiber's streaming writer for SSE
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {

		defer func() {
			// Step 4: Cleanup client connection on function exit
			ctx := context.WithValue(context.Background(), "component", "sse")
			server.LogDebug(ctx, "SSE client removed",
				"tags", []string{"notification", "network", "low"},
			)
			s.RemoveClient(userID)
		}()

		// Step 5: Setup heartbeat system for connection maintenance
		heartbeatTicker := time.NewTicker(15 * time.Second) // 15-second intervals
		defer heartbeatTicker.Stop()

		for {
			select {
			case msg, ok := <-client.Messages: // Check 'ok' to detect closed channel
				if !ok {
					// Channel closed - client was removed/replaced
					ctx := context.WithValue(context.Background(), "component", "sse")
					server.LogDebug(ctx, "SSE client channel closed",
						"tags", []string{"notification", "network", "low"},
					)
					return
				}
				// Step 9: Send notification message to client
				fmt.Fprintf(w, "data: %s\n\n", msg)
				if err := w.Flush(); err != nil {
					ctx := context.WithValue(context.Background(), "component", "sse")
					server.LogDebug(ctx, "SSE client disconnected (write error)",
						"tags", []string{"notification", "network", "low"},
					)
					return
				}

			case <-heartbeatTicker.C:
				// Step 10: Send heartbeat to maintain connection and prevent timeouts
				client.LastActive = time.Now()
				fmt.Fprintf(w, ": heartbeat\n\n")
				if err := w.Flush(); err != nil {
					ctx := context.WithValue(context.Background(), "component", "sse")
					server.LogDebug(ctx, "SSE client disconnected (heartbeat error)",
						"tags", []string{"notification", "network", "low"},
					)
					return
				}

			case <-ctxDone:
				// Step 11: Handle client disconnection (browser close, network error, etc.)
				ctx := context.WithValue(context.Background(), "component", "sse")
				server.LogDebug(ctx, "SSE client disconnected",
					"tags", []string{"notification", "network", "low"},
				)
				return
			}
		}
	})

	return nil
}

// SendNotification creates and sends a structured notification to a specific user.
// Constructs a LocalNotification with all provided metadata, serializes it to JSON,
// and delivers it through the user's SSE connection if they are connected.
//
// Parameters:
//   - userID: Target user to receive the notification (uint, required)
//   - senderID: User who triggered the notification (uint, required)
//   - entity: Type of entity - assignment, document, course, etc. (models.Entity, required)
//   - entityID: Unique identifier for the entity (uint, required)
//   - nType: Notification type - create, update, delete, etc. (notifications.NotificationType, required)
//   - title: Human-readable notification title (string, required)
//   - message: Detailed notification message (string, required)
//   - action: Action type for client-side handling (string, required)
//   - data: JSON-encoded payload with additional context (string, optional)
//
// Returns:
//   - error: JSON marshalling error, nil if successful
//
// Notification Structure:
//   - SenderID: User who triggered the notification
//   - Entity: Entity type (assignment, document, course, etc.)
//   - EntityID: Specific entity identifier
//   - Type: Notification type (create, update, delete, etc.)
//   - Action: Client action type for UI handling
//   - Title: Human-readable notification title
//   - Message: Detailed notification message
//   - Data: JSON payload with additional context
//
// Delivery Behavior:
//   - Sends to user's SSE connection if connected
//   - Graceful failure for disconnected users (not an error)
//   - Non-blocking delivery prevents server hanging
//
// Error Handling:
//   - Returns error only for JSON marshalling failures
//   - Delivery failures logged but don't return errors
//   - Disconnected users handled gracefully
//
// Side Effects:
//   - Logs marshalling errors for debugging
//   - Attempts message delivery via SendToUser
func (s *SSEServer) SendNotification(userID, senderID uint, entity models.Entity, entityID uint, nType notifications.NotificationType, title, message, action, data string) error {
	// Step 1: Construct structured notification with all provided metadata
	notification := notifications.LocalNotification{
		SenderID: senderID, // User who triggered the notification
		Entity:   entity,   // Entity type (assignment, document, etc.)
		EntityID: entityID, // Specific entity identifier
		Type:     nType,    // Notification type (create, update, etc.)
		Action:   action,   // Client action type for UI handling
		Title:    title,    // Human-readable notification title
		Message:  message,  // Detailed notification message
		Data:     data,     // JSON payload with additional context
	}

	// Step 2: Serialize notification to JSON for SSE transmission
	jsonData, err := json.Marshal(notification)
	if err != nil {
		// Step 3: Log marshalling errors for debugging
		return errors.WrapServer(err, errors.ProcJSONMarshalFailed, "Failed to marshal notification", fiber.StatusInternalServerError)
	}

	// Step 4: Attempt delivery to user's SSE connection (graceful failure if disconnected)
	s.SendToUser(userID, jsonData)

	// Step 5: Return success (delivery failures are not considered errors)
	return nil
}
