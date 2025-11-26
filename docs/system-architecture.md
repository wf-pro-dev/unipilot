# System Architecture

## Overview

UniPilot is a comprehensive course management system built with a Go backend and designed to handle course creation, sharing, and document management. The system follows a RESTful API architecture with authentication middleware and real-time notifications via Server-Sent Events (SSE).

## Architecture Components

The system is composed of several key components:

- **REST API Server**: Built with Go, handles HTTP requests and responses
- **Middleware Layer**: Comprehensive request processing with authentication, logging, and database injection
- **Authentication System**: JWT-based stateless authentication with refresh token support
- **Database Layer**: Handles data persistence for courses, assignments, and documents
- **AI Services**: Google Gemini integration for content generation and document embedding
- **Vector Database**: Qdrant integration for semantic search and RAG capabilities
- **Notification System**: Real-time notifications via dedicated SSE server (port 3000) with gRPC server (port 9000) for inter-service communication
- **Document Management**: File handling and storage for course materials

## Table of Contents

| Document | Description |
|----------|-------------|
| [API Endpoints](api-endpoints.md) | Detailed documentation of all REST API endpoints including request/response formats, authentication requirements, and status codes |
| [Database Schema](database.md) | Database table structures, relationships, Redis caching strategies, and performance considerations |
| [SSE & Notifications](sse.md) | Server-Sent Events implementation with dedicated SSE service (port 3000) and gRPC server (port 9000), real-time notification system, thread-safe client management, and inter-service communication |
| [Infrastructure](infrastructure.md) | Cloud storage architecture, RAG vector database, Docker infrastructure, and scalability considerations |

## Key Features

- **User Registration & Authentication**: Secure user account creation with JWT-based authentication
- **Course Management**: Create, update, and retrieve course information
- **Course Sharing**: Share courses between users with link-based invitations
- **Document Integration**: Associate documents with courses and assignments
- **Real-time Notifications**: Instant updates via SSE when courses are shared or accepted (dedicated service on port 3000)
- **AI-Powered Features**: Google Gemini integration for content generation and semantic search
- **Vector Search**: Qdrant database for document embeddings and RAG capabilities
- **Transaction Safety**: Database operations use transactions for data consistency
- **Caching Layer**: Redis integration for performance optimization

## Middleware Architecture

### Request Processing Pipeline
The system uses a comprehensive middleware stack for request processing:

1. **Database Middleware**: Injects database connections into request context
2. **Authentication Middleware**: Validates JWT tokens and extracts user context
3. **Logger Middleware**: Provides request tracking, performance monitoring, and audit trails

### Authentication Flow
- **JWT-based Authentication**: Stateless token-based authentication system
- **Token Types**: Access tokens (15 min) and refresh tokens (30 days)
- **User Context Injection**: Authenticated user data available to all handlers
- **Request Tracing**: Unique request IDs for distributed system debugging

### Performance Monitoring
- **Request Duration Tracking**: Millisecond-precision performance metrics
- **User Activity Logging**: Comprehensive audit trails for security and analytics
- **Structured Logging**: JSON-formatted logs for monitoring and alerting systems

## Real-time Communication Architecture

### Real-time Notification Architecture

**SSE Service Design:**
- **Dedicated HTTP Server**: Standalone server on port 3000 for SSE connections
- **Thread-Safe Operations**: RWMutex-protected client management for concurrent access
- **Connection Persistence**: Long-lived HTTP connections with 15-second heartbeat intervals
- **Message Buffering**: 100-message buffer per client to handle notification bursts

**gRPC Service Integration:**
- **Dedicated gRPC Server**: Standalone server on port 9000 for inter-service communication
- **Service Bridge**: Connects API server (HTTP) with SSE server (WebSocket-like)
- **Notification Processing**: Receives gRPC requests and forwards to SSE server
- **Health Monitoring**: Heartbeat endpoint for service discovery and load balancing

### Client Connection Management
- **Authentication Integration**: JWT-based authentication via middleware
- **Resource Management**: Automatic cleanup of disconnected clients and goroutines
- **Graceful Degradation**: Non-blocking message sends prevent slow clients from affecting others
- **Health Monitoring**: Active client count logging and connection health tracking

## Security Considerations

### Authentication & Authorization
- **JWT-based Authentication**: Stateless authentication with access and refresh token strategy
- **Token Security**: Cryptographically signed tokens with server secret validation
- **User Context**: Complete user object embedded in JWT claims for authorization
- **Session Management**: Stateless design eliminates server-side session vulnerabilities

### Data Protection
- **Password Security**: bcrypt hashing with cryptographically secure defaults
- **Database Security**: Transaction-based operations ensure data consistency and atomicity
- **Input Validation**: Comprehensive validation and sanitization to prevent injection attacks
- **User Isolation**: Database queries filtered by user context to prevent unauthorized access

### System Security
- **Request Tracing**: Unique request IDs enable security audit trails and incident response
- **Authentication Middleware**: Protects all sensitive endpoints (except public registration/login)
- **Course Sharing Security**: UUID-based sharing links prevent unauthorized course access
- **Username Uniqueness**: Prevents account conflicts and impersonation attacks