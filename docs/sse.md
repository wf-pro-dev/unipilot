# Server-Sent Events (SSE) & Real-time Notifications

## Overview

UniPilot implements a real-time notification system using Server-Sent Events (SSE) to provide instant updates to users when course-related activities occur. The system uses gRPC for internal communication between the API server and the SSE notification service.

## Architecture

### Components
- **API Server**: Triggers notifications when events occur (assignment creation, course sharing)
- **gRPC Client**: Sends notification requests to the SSE service (port 9000)
- **gRPC Server**: Receives notification requests and forwards to SSE server (port 9000)
- **SSE Service**: Manages persistent connections and delivers notifications to clients (port 3000)
- **Web Clients**: Receive real-time updates via SSE connections

### Communication Flow
1. User performs action (create assignment, share course)
2. API server processes the action and commits to database
3. API server sends notification via gRPC client to gRPC server (port 9000)
4. gRPC server receives request and forwards to SSE server
5. SSE server delivers notification to all relevant connected clients (port 3000)
6. Clients receive and display real-time updates via SSE connections

## Technical Implementation

### Server Configuration

**SSE HTTP Server:**
- **Port**: 3000 (dedicated SSE service port)
- **Endpoints**: 
  - `/health` - Public health check for load balancers and monitoring systems
  - `/unipilot/sse/v1` - Authenticated SSE endpoint with JWT middleware
- **Startup**: Non-blocking server initialization using goroutines
- **Error Handling**: Fatal errors on port conflicts or startup failures

**gRPC Server:**
- **Port**: 9000 (dedicated gRPC service port)
- **Service**: `notifications.NotificationsService`
- **Methods**: `SendNotification`, `Heartbeat`
- **Startup**: Blocking operation (runs in main goroutine)
- **Error Handling**: Fatal errors on port conflicts or startup failures

### SSE Protocol Details
- **Content-Type**: `text/event-stream` (SSE MIME type)
- **Headers**: 
  - `Cache-Control: no-cache` (prevent caching)
  - `Connection: keep-alive` (maintain persistent connection)
  - `Access-Control-Allow-Origin: *` (CORS support)
  - `X-Accel-Buffering: no` (disable proxy buffering)
- **Heartbeat**: 15-second intervals with `: heartbeat\n\n` format
- **Message Format**: `data: {json}\n\n` following SSE specification
- **Connection Validation**: HTTP flusher support required for real-time streaming

### Data Structures
```go
// SSEClient represents each connected user
type SSEClient struct {
    UserID     uint           // Authenticated user identifier
    Messages   chan []byte    // Buffered message queue (100 capacity)
    Connected  bool           // Connection status flag
    LastActive time.Time      // Heartbeat tracking
}

// LocalNotification structure sent to clients
type LocalNotification struct {
    SenderID uint   // User who triggered the notification
    Entity   string // Entity type (assignment, document, etc.)
    EntityID uint   // Specific entity identifier
    Type     string // Notification type (create, update, etc.)
    Title    string // Human-readable title
    Message  string // Detailed message
    Action   string // Client action type
    Data     string // JSON payload with additional context
}
```

### Thread Safety
- **RWMutex**: Protects client map during concurrent access
- **Read Operations**: Multiple goroutines can read client list simultaneously
- **Write Operations**: Exclusive access for adding/removing clients
- **Channel Operations**: Go channels provide built-in thread safety for message passing

## Notification Types

### Assignment Notifications

#### Assignment Creation/Update
**Trigger:** When a user creates a new assignment in a shared course

**Notification Details:**
- **Type:** `NotificationAssignmentUpdate`
- **Entity:** `EntityAssignment`
- **Recipients:** All users linked to the course (except the creator)
- **Payload:** Complete assignment data including title, deadline, course info

**Example Notification:**
```json
{
  "user_id": 123,
  "sender_id": 456,
  "entity": "assignment",
  "entity_id": 789,
  "type": "assignment_update",
  "title": "Math Homework #3",
  "message": "john_doe shared a new assignment on MATH101",
  "action": "assignment",
  "data": "{\"id\":789,\"title\":\"Math Homework #3\",\"deadline\":\"2024-01-15\",\"course_code\":\"MATH101\",...}"
}
```

**Use Cases:**
- Notify team members when new assignments are added to shared courses
- Real-time collaboration on course assignments
- Instant visibility of assignment updates across team members

### Document Notifications

#### Document Upload/Sharing
**Trigger:** When a user uploads a new document to a shared assignment

**Notification Details:**
- **Type:** `NotificationDocumentUpdate`
- **Entity:** `EntityDocument`
- **Recipients:** All users linked to the assignment's child assignments (except the uploader)
- **Payload:** Complete document metadata including filename, size, and creator info

**Example Notification:**
```json
{
  "user_id": 789,
  "sender_id": 456,
  "entity": "document",
  "entity_id": 123,
  "type": "document_update",
  "title": "Research Paper Draft",
  "message": "john_doe shared a new document on research_paper.pdf",
  "action": "document",
  "data": "{\"id\":123,\"file_name\":\"research_paper.pdf\",\"file_size\":2048576,\"assignment_id\":789,...}"
}
```

**Use Cases:**
- Notify collaborators when new documents are shared on assignments
- Real-time document sharing for team projects
- Instant visibility of file uploads across assignment participants

**Important Notes:**
- Only triggers for new file uploads (has_local_file = true)
- Does not trigger for file copying operations within the system
- Recipients determined by assignment's child assignment relationships

### Note Notifications (Prepared for Future Implementation)

#### Note Creation/Sharing
**Trigger:** When a user creates a new note in a shared course (currently disabled)

**Notification Details:**
- **Type:** `NotificationNoteUpdate`
- **Entity:** `EntityNote`
- **Recipients:** All users linked to the course (except the creator)
- **Payload:** Complete note metadata including title, subject, and content

**Example Notification (Prepared Structure):**
```json
{
  "user_id": 456,
  "sender_id": 123,
  "entity": "note",
  "entity_id": 789,
  "type": "note_update",
  "title": "Advanced Calculus Notes",
  "message": "john_doe shared a new note on MATH301",
  "action": "note",
  "data": "{\"id\":789,\"title\":\"Advanced Calculus Notes\",\"subject\":\"Derivatives\",\"course_code\":\"MATH301\",...}"
}
```

**Use Cases:**
- Notify study group members when new notes are shared
- Real-time collaboration on course materials
- Instant visibility of AI-generated content across team members

**Implementation Status:**
- **Code Prepared:** Notification logic implemented but commented out
- **Pending:** Docker deployment integration for gRPC/SSE services
- **Future Features:** Will enable real-time note sharing and collaboration

**AI Integration Notes:**
- Notifications will include information about AI-generated content
- Users will be notified when Gemini AI creates content for shared notes
- Enhanced collaboration through AI-assisted note creation

### Follow Notifications (Prepared for Future Implementation)

#### New Follower Notification
**Trigger:** When a user follows another user (currently disabled)

**Notification Details:**
- **Type:** `NotificationFollow`
- **Entity:** `EntityFollow`
- **Recipients:** The user being followed
- **Payload:** Follower information and shared course context

**Example Notification (Prepared Structure):**
```json
{
  "user_id": 456,
  "sender_id": 123,
  "entity": "follow",
  "entity_id": 456,
  "type": "follow",
  "title": "john_doe",
  "message": "john_doe followed you. You share 3 courses with this user",
  "action": "create",
  "data": ""
}
```

**Social Context Features:**
- **Shared Course Count:** Notifications include number of shared courses between users
- **Enhanced Engagement:** Rich context helps users understand connection relevance
- **Social Discovery:** Promotes course-based networking and collaboration

**Use Cases:**
- Notify users when they gain new followers
- Highlight shared academic interests through course overlap
- Encourage social engagement and course collaboration
- Build academic networking based on shared educational content

**Implementation Status:**
- **Code Prepared:** Notification logic implemented but commented out
- **Pending:** Docker deployment integration for gRPC/SSE services
- **Social Analytics:** Prepared to track follow patterns and course sharing trends

### Course Sharing Notifications

#### Course Link Acceptance
**Trigger:** When a user accepts a shared course link

**Notification Details:**
- **Recipients:** Original course owner
- **Purpose:** Confirm that their course link was accepted
- **Payload:** Information about who accepted the course

## Implementation Details

### SSE Server Architecture

**Core Components:**
- **SSEServer**: Main server managing client connections and message distribution
- **SSEClient**: Individual client connection with buffered message channel (100 messages)
- **Connection Management**: Thread-safe client registration/removal with RWMutex
- **Message Delivery**: Non-blocking sends to prevent slow clients from affecting others

**Server Initialization:**
```go
// Start SSE server on port 3000 with health check and authenticated endpoints
sseServer := NewSSEServer()
http.HandleFunc("/health", HealthHandler)
http.HandleFunc("/unipilot/sse/v1", server.AuthMiddleware(sseServer.SSEHandler))
```

**Client Connection Lifecycle:**
```go
// Add authenticated client with buffered message channel
client := s.AddClient(userID)
defer s.RemoveClient(userID) // Cleanup on disconnect

// Send notification with graceful degradation
success := s.SendToUser(userID, jsonMessage)
if !success {
    // Client disconnected or buffer full - logged but not fatal
}
```

### gRPC Integration

**Service Architecture:**
- **Port**: 9000 (dedicated gRPC service port)
- **Service**: `notifications.NotificationsService`
- **Methods**: `SendNotification`, `Heartbeat`
- **Protocol**: gRPC over TCP
- **Role**: Bridge between API server (HTTP) and SSE server (WebSocket-like)

**Service Definition:**
```go
// gRPC service implementation
type Server struct {
    UnimplementedNotificationsServiceServer
    SSE *sse.SSEServer  // Reference to SSE server for notification delivery
}
```

**gRPC Methods:**

**SendNotification Method:**
- **Purpose**: Process notification requests from API server
- **Flow**: gRPC request → type conversion → SSE server → client delivery
- **Parameters**: UserId, SenderId, Entity, EntityId, Type, Title, Message, Action, Data
- **Response**: Success status based on delivery result
- **Error Handling**: Graceful failure for disconnected users

**Heartbeat Method:**
- **Purpose**: Service health monitoring and connectivity verification
- **Use Cases**: Load balancer health checks, service discovery, latency measurement
- **Response**: Always returns "heartbeat received" confirmation

**Client Usage (from API server):**
```go
if GrpcClient != nil {
    GrpcClient.SendNotification(context.Background(), &notifications.Notification{
        UserId:   uint32(recipientID),
        SenderId: uint32(senderID),
        Entity:   string(models.EntityAssignment),
        EntityId: uint32(entityID),
        Type:     string(notif.NotificationAssignmentUpdate),
        Title:    assignmentTitle,
        Message:  fmt.Sprintf("%s shared a new assignment on %s", username, courseCode),
        Action:   "assignment",
        Data:     string(jsonData),
    })
}
```

**Server Startup:**
```go
// Initialize and start gRPC server (blocking operation)
func StartGRPCServer(sseServer *SSEServer) {
    // Creates TCP listener on port 9000
    // Registers NotificationsService
    // Starts serving gRPC requests
}
```

### Connection Management

**Authentication & Security:**
- JWT-based authentication required via AuthMiddleware
- User context extracted from validated tokens (`user_id` from JWT claims)
- Each user limited to one active SSE connection (replaces existing)
- Silent authentication failures for invalid/missing tokens

**Connection Lifecycle:**
1. JWT token validation and user extraction
2. SSE headers setup and HTTP flusher verification
3. Client registration in server connection pool
4. Message streaming with heartbeat management
5. Graceful cleanup on disconnection or context cancellation

**Connection Persistence:**
- Heartbeat every 15 seconds to prevent proxy timeouts
- LastActive timestamp tracking for connection health
- Graceful handling of client disconnections via context cancellation
- Automatic resource cleanup (channels closed, clients removed)

**Message Buffering:**
- 100-message buffer per client to handle burst notifications
- Non-blocking sends prevent server blocking on slow clients
- Buffer overflow results in dropped messages (graceful degradation)
- Thread-safe operations using RWMutex for concurrent access

### Error Handling

**Graceful Degradation:**
- gRPC client availability is checked before sending notifications
- Notification failures are logged but don't affect main operation success
- System continues to function normally even if SSE service is unavailable
- Message buffer overflow silently drops messages to prevent memory issues

**Connection Resilience:**
- Automatic client cleanup on disconnection prevents resource leaks
- Failed message sends don't crash the server or affect other clients
- Heartbeat mechanism detects and handles stale connections

**Logging:**
- All notification attempts are logged for debugging
- Failed notifications are logged as warnings with full context
- Performance metrics include notification delivery timing
- Active client count logged for monitoring connection health

## Security Considerations

### Access Control
- Users only receive notifications for courses they have access to
- Sender identity is verified through JWT authentication
- Notification recipients are determined by course membership

### Data Privacy
- Notification payloads only include data the recipient has permission to see
- Sensitive user information is excluded from notification messages
- Course sharing respects existing permission models

## Performance Considerations

### Scalability
- gRPC provides efficient binary communication for high-throughput scenarios
- SSE connections are managed by dedicated service for optimal resource usage
- Notification delivery is asynchronous and non-blocking

### Resource Management
- Failed notification attempts don't retry automatically to prevent cascading failures
- Connection management handled by SSE service with appropriate timeouts
- Memory usage optimized through efficient message serialization
- Buffered channels (100 messages) prevent memory buildup during client slowdowns
- Automatic goroutine cleanup when clients disconnect prevents resource leaks
- Thread-safe operations using RWMutex for optimal concurrent performance

## Monitoring & Debugging

### Logging Strategy
- All notification events logged with structured data
- Request IDs maintained across service boundaries for tracing
- Performance metrics captured for notification delivery times

### Health Checks
- gRPC client connection status monitored
- SSE service availability checked through health endpoints
- Notification delivery success rates tracked for system health

## Future Enhancements

### Planned Features
- Message queuing for guaranteed delivery during service outages
- Notification preferences and filtering for users
- Real-time typing indicators for collaborative editing
- Connection pooling for improved resource utilization

### Scalability Improvements
- Horizontal scaling of SSE service instances
- Load balancing for notification distribution
- Caching layer for frequently accessed notification data
- Redis-based message broadcasting for multi-instance deployments
