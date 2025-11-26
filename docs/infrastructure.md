# Infrastructure & Cloud Services

## Overview

UniPilot leverages a hybrid cloud infrastructure combining traditional database storage with modern cloud services and AI-powered vector databases. The system is designed for scalability, performance, and advanced document processing capabilities.

## Cloud Storage Architecture

### AWS S3 Integration
**Purpose:** Primary file storage for documents, assignments, and user-generated content

**Storage Organization:**
```
users_data/
├── user_{id}/
│   └── documents/
│       └── assign_{assignment_id}/
│           ├── {timestamp}_{filename1}
│           ├── {timestamp}_{filename2}
│           └── ...
```

**Key Features:**
- **Hierarchical Organization:** Files organized by user and assignment for easy management
- **Unique Naming:** Timestamp-based prefixes prevent filename conflicts
- **Scalable Storage:** Leverages AWS S3's virtually unlimited storage capacity
- **Cost Optimization:** Pay-per-use model scales with actual usage

**Security Measures:**
- **Access Control:** Storage keys used instead of direct file paths
- **User Isolation:** Directory structure prevents cross-user file access
- **Secure Transfer:** Files uploaded and downloaded through authenticated API endpoints

### File Processing Pipeline

**Upload Process:**
1. **Multipart Form Parsing:** Handle large files with memory-efficient streaming
2. **Temporary Local Storage:** Files written to `/app/uploads` Docker volume
3. **S3 Upload:** Files transferred to AWS S3 with organized naming
4. **Local Cleanup:** Temporary files removed after successful S3 upload
5. **Metadata Storage:** File information stored in PostgreSQL database

**Download Process:**
1. **Storage Key Lookup:** Retrieve S3 key from database metadata
2. **Direct Streaming:** Files streamed from S3 to client without local storage
3. **Proper Headers:** Content-Type and Content-Disposition set for secure downloads

## Vector Database (RAG) Infrastructure

### Qdrant Vector Database
**Purpose:** Stores AI-generated embeddings for semantic document search and retrieval

**Architecture:**
- **Collection-Based:** Separate collections per assignment for data isolation
- **Vector Dimensions:** 768-dimensional embeddings for rich semantic representation
- **Similarity Metric:** Cosine distance for accurate semantic matching
- **Scalable Storage:** Handles large document collections with efficient indexing

**Collection Management:**
```
Collection Naming: unipilot-qdrant-db-{assignment_id}
Vector Config:
  - Size: 768 dimensions
  - Distance: Cosine similarity
  - Indexing: Automatic for fast retrieval
```

**RAG Processing Pipeline:**
1. **Document Ingestion:** Files processed from S3 or local uploads
2. **Text Extraction:** Content extracted from various file formats
3. **Chunking:** Documents split into manageable segments
4. **Vectorization:** AI embeddings generated for each chunk
5. **Storage:** Vectors stored in assignment-specific Qdrant collections
6. **Indexing:** Automatic indexing for fast semantic search

### AI Integration
**Embedding Generation:**
- **Model:** Advanced language models for high-quality embeddings
- **Processing:** Batch processing for efficiency
- **Quality:** 768-dimensional vectors capture rich semantic meaning

**Search Capabilities:**
- **Semantic Search:** Find documents by meaning, not just keywords
- **Context Retrieval:** Retrieve relevant document chunks for AI responses
- **Assignment Scope:** Search limited to assignment-specific collections

## AI Services Integration

### Google Gemini Integration
**Purpose:** Provides AI-powered content generation for educational notes and materials

**Note Generation Pipeline:**
1. **Context Analysis:** Uses course code, title, and subject as generation context
2. **Content Generation:** Creates relevant educational content using Gemini AI
3. **Keyword Extraction:** Generates appropriate keywords for search and categorization
4. **Quality Assurance:** Validates generated content for educational relevance

**API Integration:**
- **Service:** Google Gemini AI API
- **Request Structure:** Title, subject, course name for context
- **Response Processing:** Content and keywords extraction
- **Error Handling:** Graceful fallback to manual content creation

**Content Quality Features:**
- **Educational Focus:** Content tailored for academic and learning contexts
- **Context Awareness:** Generation considers course subject and level
- **Keyword Optimization:** Automatic keyword generation for better searchability
- **Customization:** User can override AI-generated content with manual input

**Performance Considerations:**
- **Async Processing:** AI generation doesn't block note creation workflow
- **Caching Strategy:** Consider caching common educational topics
- **Rate Limiting:** Respect API rate limits for sustainable usage
- **Fallback Mechanisms:** Manual content creation when AI service unavailable

### AI Service Architecture (Node.js)

**Purpose:** Comprehensive AI service providing embedding generation, vector search, and RAG capabilities

**Service Components:**

#### 1. AI Embedding Service (`lib/embedding.js`)
- **Technology:** Google Gemini AI (gemini-embedding-001 model)
- **Function:** Converts text documents into 768-dimensional semantic vectors
- **Features:** Text preprocessing, error handling, and optimized vector generation
- **Integration:** Powers document vectorization for semantic search capabilities

#### 2. Vector Database Service (`lib/qdrant.js`)
- **Technology:** Qdrant vector database with gRPC client
- **Function:** Manages vector collections, similarity search, and RAG operations
- **Features:** Collection management, semantic search, and content retrieval
- **Integration:** Stores and queries document embeddings for intelligent search

**Core Functions:**
- `ListCollections()`: Administrative function for collection management and monitoring
- `findRelevantContent(userQuery, assignmentID)`: Complete RAG pipeline implementation

**RAG Implementation Details:**
1. **Query Processing:** Converts natural language queries to 768-dimensional embeddings
2. **Vector Search:** Performs semantic similarity search in assignment-specific collections
3. **Content Retrieval:** Returns top 3 most relevant document chunks with relevance scores
4. **AI Response Generation:** Uses Google Gemini to synthesize contextual answers
5. **Error Handling:** Comprehensive error management with graceful degradation

**Collection Architecture:**
- **Naming Convention:** `unipilot-qdrant-db-{assignmentID}` for data isolation
- **Payload Structure:** Includes `chunk_text` and `assignment_id` metadata
- **Search Configuration:** Cosine similarity with configurable result limits
- **Performance:** Sub-second search times with automatic indexing

#### 3. RAG API Server (`index.js`)
- **Technology:** Express.js REST API with Google Gemini integration
- **Function:** Provides retrieval-augmented generation endpoints
- **Features:** Streaming text generation, tool integration, and CORS support
- **Integration:** Serves AI-powered document queries and content generation

**AI Processing Pipeline:**
1. **Document Ingestion:** Text documents processed through embedding service
2. **Vector Generation:** 768-dimensional embeddings created using Gemini AI
3. **Vector Storage:** Embeddings stored in assignment-specific Qdrant collections
4. **Semantic Search:** Query vectors matched against document embeddings
5. **RAG Generation:** Retrieved content used for context-aware AI responses

**Technology Stack:**
- **Runtime:** Node.js with Express.js framework
- **AI Provider:** Google Gemini AI for embeddings and text generation
- **Vector Database:** Qdrant with gRPC client for high-performance operations
- **SDK Integration:** AI SDK with Google provider and streaming capabilities

**Service Architecture:**
- **Microservice Design:** Dedicated Node.js service for AI operations
- **API Integration:** RESTful endpoints for Go backend communication
- **Async Processing:** Non-blocking AI operations with streaming responses
- **Error Resilience:** Comprehensive error handling and graceful degradation

**Performance Optimizations:**
- **Batch Processing:** Efficient handling of multiple document embeddings
- **Streaming Responses:** Real-time AI text generation with streaming
- **Connection Pooling:** Optimized Qdrant client connections
- **Rate Limiting:** Intelligent API quota management for Google AI services

**Security & Compliance:**
- **API Key Security:** Environment-based credential management
- **CORS Configuration:** Controlled cross-origin access for web clients
- **Data Privacy:** Secure handling of document content through AI services
- **Service Authentication:** Secure communication between Go backend and AI service

## Docker Infrastructure

### Container Architecture
**Application Container:**
- **Base:** Go application with all dependencies
- **Volumes:** `/app/uploads` for temporary file processing
- **Networking:** Internal communication with database and external services

**Volume Management:**
- **Uploads Volume:** Temporary storage for file processing
- **Persistent Data:** Database and configuration data preserved
- **Cleanup Strategy:** Temporary files automatically removed after processing

### Service Dependencies
**Required Services:**
- **PostgreSQL:** Primary database for metadata and relationships
- **Redis:** Caching layer for performance optimization
- **AWS S3:** Cloud storage for files and documents
- **Qdrant:** Vector database for RAG functionality
- **gRPC Service:** Real-time notifications and SSE

## Performance Considerations

### Storage Performance
**S3 Optimization:**
- **Multipart Uploads:** Efficient handling of large files
- **Direct Streaming:** No intermediate storage for downloads
- **Parallel Processing:** Multiple file operations can run concurrently

**Local Storage:**
- **Docker Volumes:** Fast local storage for temporary processing
- **Automatic Cleanup:** Prevents disk space accumulation
- **Memory Management:** 32MB multipart form limit prevents memory exhaustion

### Vector Database Performance
**Qdrant Optimization:**
- **Collection Isolation:** Separate collections improve query performance
- **Indexing Strategy:** Automatic indexing for sub-second search times
- **Memory Usage:** Efficient vector storage and retrieval

## Monitoring & Observability

### Storage Monitoring
**Metrics Tracked:**
- **Upload Success Rates:** Monitor file upload reliability
- **Storage Usage:** Track per-user and total storage consumption
- **Transfer Speeds:** Monitor upload/download performance
- **Error Rates:** Track storage operation failures

**Logging Strategy:**
- **Structured Logging:** JSON format for easy parsing
- **Request Tracing:** Unique request IDs for operation tracking
- **Performance Metrics:** Duration tracking for all storage operations

### RAG System Monitoring
**Vector Database Health:**
- **Collection Status:** Monitor collection creation and health
- **Vector Count:** Track document processing volume
- **Query Performance:** Monitor search response times
- **Indexing Status:** Ensure proper vector indexing

## Security & Compliance

### Authentication Architecture

**JWT-Based Authentication:**
- **Stateless Design:** No server-side session storage required
- **Token Structure:** Complete user object embedded in JWT claims
- **Dual Token System:** Access tokens (15 min) + Refresh tokens (30 days)
- **Signature Validation:** HMAC-SHA256 with server secret key

**Middleware Security Stack:**
- **Authentication Middleware:** JWT validation and user context injection
- **Database Middleware:** Secure database connection management
- **Logger Middleware:** Comprehensive audit trail generation

**Request Security:**
- **Unique Request IDs:** UUID-based tracing for security analysis
- **User Context Validation:** Every request validated against user permissions
- **Performance Monitoring:** Request duration tracking for anomaly detection

### Data Protection
**Encryption:**
- **In Transit:** HTTPS for all API communications with TLS 1.2+
- **At Rest:** AWS S3 server-side encryption with AES-256
- **Vector Data:** Secure storage in Qdrant collections with access controls
- **JWT Tokens:** Cryptographically signed with server secret

**Access Control:**
- **Authentication:** JWT-based stateless user authentication
- **Authorization:** User-specific file access controls and data isolation
- **Database Security:** User-scoped queries prevent unauthorized data access
- **Assignment Isolation:** Vector collections separated by assignment for data privacy

### Security Monitoring

**Request Tracking:**
- **Audit Trails:** Comprehensive logging of all user actions and system events
- **Performance Metrics:** Request duration and volume monitoring for anomaly detection
- **User Activity:** Detailed tracking of authentication events and data access patterns
- **Security Events:** Failed authentication attempts and suspicious activity logging

**Compliance Features:**
- **Data Retention:** Configurable log retention policies for compliance requirements
- **User Privacy:** GDPR-compliant data handling and user data export capabilities
- **Access Logs:** Detailed audit trails for security compliance and incident response
- **Encryption Standards:** Industry-standard encryption for data protection compliance

### Backup & Recovery
**Data Backup:**
- **Database:** Regular PostgreSQL backups with encryption at rest
- **File Storage:** AWS S3 built-in durability (99.999999999%) with versioning
- **Vector Data:** Qdrant collection backup strategies with point-in-time recovery
- **Security Logs:** Audit trail backups for compliance and forensic analysis

**Disaster Recovery:**
- **Multi-Region:** AWS S3 cross-region replication for geographic redundancy
- **Database Recovery:** Point-in-time recovery with encrypted backup validation
- **Service Redundancy:** Multiple service instance deployment with load balancing
- **Security Continuity:** Authentication service redundancy and failover capabilities

## Scalability Planning

### Horizontal Scaling
**Service Scaling:**
- **Application Instances:** Multiple Go application containers
- **Load Balancing:** Distribute requests across instances
- **Database Scaling:** Read replicas for query performance

**Storage Scaling:**
- **S3 Auto-Scaling:** Automatic capacity management
- **Qdrant Clustering:** Multi-node vector database deployment
- **Cache Scaling:** Redis cluster for increased cache capacity

### Performance Optimization
**Future Enhancements:**
- **CDN Integration:** CloudFront for global file distribution
- **Caching Layers:** Multi-level caching for frequently accessed files
- **Compression:** File compression for storage and transfer efficiency
- **Batch Processing:** Bulk operations for improved throughput
