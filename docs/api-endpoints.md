# REST API Endpoints

## User Management

### POST `/unipilot/api/v1/register`
**Purpose:** Registers a new user account in the system.

**Description:** Creates a new user with hashed password, generates JWT access and refresh tokens, and caches user data in Redis for performance optimization. Uses bcrypt for secure password hashing.

**Request Body:**
- `username`: User's chosen username (string, required)
- `email`: User's email address (string, required)
- `password`: User's password in plain text (string, required)
- `university`: User's university affiliation (string, required)
- `language`: User's preferred language (string, required)

**Response:**
- `message`: Success message
- `user`: User object (as map) with sensitive fields removed
- `token`: JWT access token (expires in 15 minutes)
- `refresh_token`: JWT refresh token (expires in 30 days)

**Authentication:** None required (public endpoint)

**Database Operations:**
- Checks for existing username in `users` table
- Creates new user record in `users` table
- Caches user data in Redis (non-blocking operation)

**Security Features:**
- Password hashing using bcrypt with default cost
- JWT tokens with appropriate expiration times
- Username uniqueness validation

**Status Codes:**
- 200 OK: User registered successfully
- 400 Bad Request: Invalid JSON body or missing required fields
- 405 Method Not Allowed: Non-POST request
- 409 Conflict: Username already exists
- 500 Internal Server Error: Password hashing, database, or token generation failure

---

### POST `/unipilot/api/v1/login`
**Purpose:** Authenticates existing users and provides access tokens.

**Description:** Validates user credentials against stored password hash using bcrypt for secure comparison. Generates new JWT access and refresh tokens for authenticated sessions with appropriate expiration times.

**Request Body:**
- `username`: User's username (string, required)
- `password`: User's password in plain text (string, required)

**Response:**
- `message`: Success message
- `user`: User object (as map) with sensitive fields removed
- `token`: JWT access token (expires in 15 minutes)
- `refresh_token`: JWT refresh token (expires in 30 days)

**Authentication:** None required (public endpoint)

**Database Operations:**
- Queries `users` table by username
- Retrieves user record with password hash for verification

**Security Features:**
- Constant-time password comparison using bcrypt
- JWT tokens with secure signing and expiration
- Structured logging for security audit trails
- No password information in response or logs

**Status Codes:**
- 200 OK: Authentication successful
- 400 Bad Request: Invalid JSON body
- 401 Unauthorized: User not found or invalid password
- 405 Method Not Allowed: Non-POST request
- 500 Internal Server Error: Session key retrieval or token generation failure

---

### POST `/unipilot/api/v1/logout`
**Purpose:** Handles user logout requests with stateless JWT token approach.

**Description:** Provides a logout acknowledgment for client-side session termination. Since the system uses stateless JWT tokens, this endpoint serves as a logout confirmation rather than server-side session invalidation. Clients must discard tokens after receiving this response.

**Request Body:** None required

**Response:**
- `message`: Logout confirmation message

**Authentication:** None required (public endpoint)

**Database Operations:** None

**Security Implementation:**
- Stateless logout design (no server-side token invalidation)
- JWT tokens remain valid until natural expiration
- Client-side token disposal responsibility
- Audit logging for security monitoring

**Status Codes:**
- 200 OK: Logout acknowledged successfully

**Important Notes:**
- This is a client-side logout implementation
- JWT access tokens remain valid for 15 minutes after logout
- JWT refresh tokens remain valid for 30 days after logout
- For enhanced security, consider implementing server-side token blacklisting
- Client applications should immediately discard all tokens

---

### POST `/unipilot/api/v1/token/refresh`
**Purpose:** Refreshes JWT tokens to maintain continuous session access.

**Description:** Generates new access and refresh tokens for authenticated users without requiring re-authentication. Uses the existing user context from a valid refresh token to create completely new tokens with fresh expiration times. This enables seamless session continuation.

**Request Body:** None required (user context extracted from existing refresh token)

**Response:**
- `message`: Success message
- `token`: New JWT access token (expires in 15 minutes)
- `refresh_token`: New JWT refresh token (expires in 30 days)

**Authentication:** Required (AuthMiddleware) - validates existing refresh token

**Database Operations:** None (purely token-based operation)

**Security Features:**
- Generates completely new tokens rather than extending expiration
- Maintains same user context with fresh timestamps
- Uses secure JWT signing with HS256 algorithm
- Appropriate token lifespans for security and usability
- Previous tokens remain valid until natural expiration

**Status Codes:**
- 200 OK: Tokens refreshed successfully
- 401 Unauthorized: Invalid or expired refresh token
- 500 Internal Server Error: Session key retrieval or token generation failure

**Important Notes:**
- Requires valid refresh token in Authorization header
- Creates entirely new token pair (not just expiration extension)
- Previous tokens continue to work until they naturally expire
- Essential for maintaining long-term sessions without re-authentication
- Should be called before access token expiration (within 15 minutes)

---

### GET `/unipilot/api/v1/user`
**Purpose:** Retrieves the current authenticated user's profile information.

**Description:** Returns user data from the JWT token context without requiring database queries. Uses cached user information from the authentication middleware for optimal performance. Sensitive fields are automatically excluded from the response.

**Request Body:** None required (user context extracted from JWT token)

**Response:**
- `message`: Success message
- `user`: User object (as map) with sensitive fields removed

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:** None (uses cached context data from authentication)

**Security Features:**
- Automatic exclusion of sensitive fields (password hash, etc.)
- No database exposure or additional queries
- Safe serialization with error handling
- Context-based user identification

**Status Codes:**
- 200 OK: User information retrieved successfully
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: User data processing failure

---

### POST `/unipilot/api/v1/user/update`
**Purpose:** Updates a specific field of the authenticated user's profile.

**Description:** Performs targeted field updates using raw SQL for flexibility. Automatically updates the `updated_at` timestamp and refreshes Redis cache for performance optimization. Only the authenticated user can update their own profile.

**Request Body:**
- `column`: Database column name to update (string, required)
- `value`: New value for the specified column (string, required)

**Response:**
- `message`: Success message
- `user`: Updated user object (as map) with sensitive fields removed

**Authentication:** Required (AuthMiddleware) - needs user context for identification

**Database Operations:**
- Updates specified column in `users` table using raw SQL
- Automatically updates `updated_at` timestamp
- Retrieves updated user record for response verification
- Updates Redis cache with new user data (non-blocking)

**Security Features:**
- User can only update their own profile (context-based authorization)
- Sensitive fields excluded from response
- Automatic timestamp tracking for audit trails
- Redis cache synchronization

**Status Codes:**
- 200 OK: User updated successfully
- 400 Bad Request: Invalid JSON body or malformed request
- 401 Unauthorized: Invalid or missing JWT token
- 404 Not Found: User record not found in database
- 500 Internal Server Error: Database operations or serialization failure

**Security Warning:** Uses string interpolation for column name - validate input to prevent SQL injection attacks.

---

### GET `/unipilot/api/v1/users`
**Purpose:** Retrieves all users in the system except the current authenticated user.

**Description:** Returns a list of all users with their associated course codes. Implements a Redis caching strategy for performance optimization - first attempts to retrieve from cache, falls back to database query if cache miss occurs. Results are cached in Redis with 1-hour TTL.

**Request Body:** None required (user context extracted from JWT token)

**Response:**
- `message`: Success message
- `users`: Array of user objects (as maps) with course codes included

**Authentication:** Required (AuthMiddleware) - extracts current user from JWT token

**Database Operations:**
- Reads from `users` table excluding current user (`WHERE id != ?`)
- Queries `courses` table to get course codes for each user
- Caches results in Redis hash with user ID as key

**Caching Strategy:**
- Redis key: `users` (hash structure)
- Hash field: User ID (as string)
- Hash value: JSON-serialized user object with course codes
- TTL: 1 hour
- Cache hit: Returns cached data directly
- Cache miss: Queries database, populates cache, returns fresh data

**Security Features:**
- Current user excluded from results (prevents self-disclosure)
- User context validation through AuthMiddleware
- Sensitive fields excluded from user objects

**Status Codes:**
- 200 OK: Users retrieved successfully (from cache or database)
- 500 Internal Server Error: Database query, Redis operations, or JSON serialization failure

**Performance Notes:**
- Cache-first strategy reduces database load
- Individual user caching allows partial cache updates
- Non-blocking cache operations (warnings logged on failure)

---

## Assignment Management

### GET `/unipilot/api/v1/assignments`
**Purpose:** Retrieves all assignments belonging to the authenticated user.

**Description:** Returns a comprehensive list of user's assignments with all assignment details converted to map format for consistent JSON serialization. Provides data for task management interface with performance tracking.

**Request Body:** None required (user context extracted from JWT token)

**Response:**
- `message`: Success message
- `assignments`: Array of assignment objects (as maps) with complete assignment details

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Reads from `assignments` table filtered by `user_id`
- Direct database access without caching strategy
- Uses ToMap() method for safe JSON serialization

**Security Features:**
- User can only access their own assignments (user_id filtering)
- Request tracking with unique request ID for audit trail
- Safe JSON serialization excluding sensitive fields

**Status Codes:**
- 200 OK: Assignments retrieved successfully
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Database query failure

**Performance Notes:**
- Request duration tracking for performance monitoring
- No caching implemented (direct database queries)
- Optimized for real-time assignment data

---

### POST `/unipilot/api/v1/assignments`
**Purpose:** Creates a new assignment for the authenticated user with real-time notifications.

**Description:** Validates input data, creates assignment record using database transactions for atomicity, and sends real-time SSE notifications to all users linked to the associated course. Supports hierarchical assignments through parent-child relationships.

**Request Body:**
- `local_id`: Assignment local identifier (string, required)
- `title`: Assignment title (string, required)
- `todo`: Assignment description/tasks (string, optional)
- `deadline`: Assignment deadline in YYYY-MM-DD format (string, required)
- `course_code`: Associated course code (string, required)
- `type`: Assignment type/category (string, required)
- `status`: Assignment status (string, optional)
- `priority`: Assignment priority level (string, optional)
- `link`: Related link/URL (string, optional)
- `parent_id`: Parent assignment ID for sub-assignments (string, optional)

**Response:**
- `message`: Success message
- `assignment`: Created assignment object (as map) with all details

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Creates record in `assignments` table using transaction
- Validates course association and user permissions
- Retrieves complete assignment data with relationships
- Transaction rollback on any failure for data consistency

**Notification System:**
- Sends SSE notifications to all users linked to the course
- Uses gRPC client for real-time notification delivery
- Notification includes assignment data and sender information
- Notification type: `NotificationAssignmentUpdate`

**Security Features:**
- Input validation for required fields and data formats
- User isolation (assignments belong to authenticated user)
- Transaction-based operations for data integrity

**Status Codes:**
- 200 OK: Assignment created successfully
- 400 Bad Request: Invalid JSON, missing fields, or format errors
- 401 Unauthorized: Invalid or missing JWT token
- 409 Conflict: Assignment creation failure (constraint violations)
- 500 Internal Server Error: Database operations or notification failures

**Side Effects:**
- Creates assignment record in database
- Sends real-time notifications via SSE to course participants
- Logs creation with performance metrics and assignment details

---

### POST `/unipilot/api/v1/assignments/update`
**Purpose:** Updates a specific field of an existing assignment with ownership validation.

**Description:** Performs targeted field updates using raw SQL within a database transaction for data consistency. Only the assignment owner can update their assignments. Uses automatic timestamp tracking for audit trails.

**Request Body:**
- `id`: Assignment ID to update (string, required, converted to int)
- `column`: Database column name to update (string, required)
- `value`: New value for the specified column (string, required)

**Response:** 200 OK with no response body (success indicated by status code)

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Validates assignment ownership by user_id before updates
- Updates specified column in `assignments` table using raw SQL
- Automatically updates `updated_at` timestamp
- Uses database transaction for atomicity

**Security Features:**
- Assignment ownership verification before allowing updates
- Transaction rollback on any failure for data consistency
- Input validation for assignment ID format

**Security Warning:** Uses string interpolation for column name - column should be validated to prevent SQL injection attacks.

**Status Codes:**
- 200 OK: Assignment updated successfully
- 400 Bad Request: Invalid JSON body, malformed assignment ID
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Database operations failure

**Side Effects:**
- Modifies assignment record in database with timestamp update
- Logs successful updates with change details for audit trail
- Request duration tracking for performance monitoring

---

## Document Management

### POST `/unipilot/api/v1/documents`
**Purpose:** Creates a new document record with file upload to cloud storage and real-time notifications.

**Description:** Handles multipart form uploads, processes document metadata, stores files in AWS S3, and sends real-time notifications to users linked to the associated assignment. Supports both new file uploads and copying existing files within the storage system.

**Request Body (multipart form):**
- `metadata`: JSON string containing LocalDocument structure (required)
- `file`: Binary file data (required if has_local_file is true)

**LocalDocument metadata fields:**
- `id`: Local document identifier (uint, required)
- `assignment_id`: Local assignment ID (uint, required)
- `remote_assignment_id`: Server assignment ID (uint, required)
- `file_name`: Original file name (string, required)
- `file_type`: MIME type (string, required)
- `file_size`: File size in bytes (int64, required)
- `type`: Document type/category (string, required)
- `version`: Document version number (int, required)
- `is_original`: Whether this is the original document (bool, required)
- `has_local_file`: Whether file data is included in request (bool, required)
- `storage_key`: Existing storage key for file copying (string, optional)

**Response:**
- `success`: Boolean success indicator
- `document`: Created document object with metadata

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Cloud Storage Operations:**
- Uploads new files to AWS S3 with unique timestamped names
- Copies existing files within S3 storage system
- Organizes files in user/assignment directory structure: `users_data/user_{id}/documents/assign_{id}/`
- Cleans up temporary local files after S3 upload

**Notification System:**
- Sends SSE notifications to users linked to assignment's child assignments
- Uses gRPC client for real-time notification delivery
- Notification type: `NotificationDocumentUpdate`
- Only sends notifications for new file uploads (not copies)

**Security Features:**
- User isolation through directory structure (user_id based paths)
- File size limits enforced through multipart form parsing (32MB)
- Unique file naming prevents conflicts and overwrites
- Assignment ownership validation through linked assignments

**Status Codes:**
- 200 OK: Document created successfully
- 400 Bad Request: Invalid multipart form, missing metadata, or format errors
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Storage operations, database failures, or notification errors

---

### POST `/unipilot/api/v1/documents/download`
**Purpose:** Streams document files directly from AWS S3 to client with secure access.

**Description:** Provides secure file download with proper headers and streaming for large files. Uses cloud storage service to retrieve files without storing them locally, enabling efficient file distribution.

**Request Body:**
- `storage_key`: S3 storage key for the file to download (string, required)
- `file_name`: Original file name for download headers (string, required)
- `file_size`: File size for Content-Length header (int64, required)

**Response:** Binary file stream with appropriate headers
- `Content-Type`: application/octet-stream
- `Content-Disposition`: attachment with original filename
- `Content-Length`: File size for download progress

**Authentication:** Not explicitly required (public endpoint for file access)

**Cloud Storage Operations:**
- Downloads file from AWS S3 using storage key
- Streams file directly to HTTP response without local storage
- Handles large files efficiently through streaming

**Security Features:**
- Uses storage keys for file identification (not direct file paths)
- Proper Content-Disposition headers prevent XSS attacks
- Streaming prevents memory exhaustion on large files

**Status Codes:**
- 200 OK: File streamed successfully
- 400 Bad Request: Invalid JSON body or missing required fields
- 500 Internal Server Error: S3 download failures or streaming errors

---

### GET `/unipilot/api/v1/documents/assignment`
**Purpose:** Retrieves all document metadata for a specific assignment with user-specific access flags.

**Description:** Returns comprehensive document information with user-specific flags for local file availability. Provides document listings for assignment management interfaces with proper access control and ownership indicators.

**Query Parameters:**
- `assignment_id`: Assignment ID to retrieve documents for (string, required)

**Response:**
- `success`: Boolean success indicator
- `documents`: Array of DocumentMetadata objects with comprehensive document information

**DocumentMetadata includes:**
- `id`, `assignment_id`, `user_id`: Database identifiers
- `type`, `file_name`, `file_type`, `file_size`: File information
- `version`, `is_original`: Version control information
- `has_local_file`: Whether current user owns this document
- `created_at`: Document creation timestamp

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Queries `documents` table filtered by assignment_id
- Preloads user information for document ownership details
- Orders results by creation date (newest first)

**Security Features:**
- Assignment ID validation and conversion
- User-specific has_local_file flag (only true for document owner)
- Safe metadata exposure without sensitive internal fields

**Status Codes:**
- 200 OK: Documents retrieved successfully
- 400 Bad Request: Missing or invalid assignment_id parameter
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Database query failures

---

### DELETE `/unipilot/api/v1/documents`
**Purpose:** Removes document record and associated file from cloud storage with ownership validation.

**Description:** Provides secure document deletion with ownership validation and storage cleanup. Updates user storage quota information after successful deletion to maintain accurate storage tracking.

**Query Parameters:**
- `document_id`: Local document ID to delete (string, required)

**Response:**
- `success`: Boolean success indicator
- `message`: Confirmation message

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Queries document by local_id and user_id for ownership validation
- Deletes document record from database
- Updates user storage quota information

**Cloud Storage Operations:**
- Deletes associated file from AWS S3 using file path
- Handles storage cleanup to prevent orphaned files

**Security Features:**
- Document ownership validation (user can only delete their own documents)
- Uses local_id and user_id combination for secure identification
- Prevents unauthorized deletion of other users' documents

**Status Codes:**
- 200 OK: Document deleted successfully
- 400 Bad Request: Missing document_id parameter
- 401 Unauthorized: Invalid or missing JWT token
- 404 Not Found: Document not found or not owned by user
- 500 Internal Server Error: Storage deletion or database failures

---

### POST `/unipilot/api/v1/documents/rag`
**Purpose:** Processes documents for Retrieval-Augmented Generation (RAG) with AI embeddings and vector storage.

**Description:** Handles document upload/download, converts to vectors using AI embeddings, and stores in Qdrant vector database for semantic search and AI-powered document retrieval. Creates assignment-specific vector collections for organized AI-powered search.

**Request Body (multipart form):**
- `metadata`: JSON string containing LocalDocument structure (required)
- `file`: Binary file data (required if has_local_file is true)

**Response:**
- `success`: Boolean success indicator
- `document`: Processed document object with RAG metadata

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**RAG Processing Pipeline:**
1. File preparation (upload new or download existing from S3)
2. Document vectorization using AI embeddings (768-dimensional vectors)
3. Qdrant collection creation (per-assignment collections)
4. Vector storage in Qdrant database for semantic search

**Qdrant Integration:**
- Creates collections named "unipilot-qdrant-db-{assignment_id}"
- Uses 768-dimensional vectors with cosine distance similarity
- Stores document chunks as searchable vectors for AI retrieval
- Handles collection creation and vector upsertion

**File Processing:**
- Supports both new uploads and existing file processing
- Downloads existing files from S3 for vectorization
- Processes documents into embeddings for semantic search
- Maintains file organization in local storage during processing

**Security Features:**
- User isolation through assignment-based collections
- File access validation through existing document records
- Secure vector storage with assignment-level separation

**Status Codes:**
- 200 OK: Document processed for RAG successfully
- 400 Bad Request: Invalid multipart form, missing metadata, or format errors
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: File processing, vectorization, or Qdrant failures

---

## Note Management

### GET `/unipilot/api/v1/notes`
**Purpose:** Retrieves all notes belonging to the authenticated user for note management interface.

**Description:** Returns a comprehensive list of user's notes with all note details converted to map format for consistent JSON serialization. Provides data for personal note management and study organization interfaces.

**Request Body:** None required (user context extracted from JWT token)

**Response:**
- `message`: Success message
- `notes`: Array of note objects (as maps) with complete note details

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Reads from `notes` table filtered by `user_id`
- Direct database access without caching strategy
- Uses ToMap() method for safe JSON serialization

**Security Features:**
- User can only access their own notes (user_id filtering)
- Safe JSON serialization excluding sensitive fields
- No cross-user data exposure

**Status Codes:**
- 200 OK: Notes retrieved successfully
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Database query failure

**Performance Notes:**
- No caching implemented (direct database queries)
- Optimized for real-time note data access

---

### POST `/unipilot/api/v1/notes`
**Purpose:** Creates a new note with AI-powered content generation using Google Gemini.

**Description:** Validates input data, optionally generates content using Google Gemini AI, creates note record using database transactions, and prepares for future notification distribution to linked users. Provides intelligent note scaffolding for educational content.

**Request Body:**
- `user_id`: User identifier (string, optional - overridden by JWT context)
- `course_code`: Associated course code (string, required)
- `title`: Note title (string, required)
- `subject`: Note subject/topic (string, required)
- `content`: Note content (string, optional - generated by AI if empty)
- `keywords`: Note keywords (string, optional - generated by AI if empty)
- `videos`: Associated video links (string, optional)

**Response:**
- `message`: Success message
- `note`: Created note object (as map) with all details including AI-generated content

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**AI Integration (Google Gemini):**
- Automatically generates content and keywords if both are empty in request
- Uses title, subject, and course code as context for AI generation
- Provides intelligent note scaffolding for educational content
- Handles AI service failures gracefully with error responses

**Database Operations:**
- Creates record in `notes` table using transaction for atomicity
- Validates note ownership and course association
- Retrieves complete note data with relationships for response

**Notification System (Prepared):**
- Code prepared for SSE notifications to course-linked users
- Currently disabled pending Docker deployment
- Will notify collaborators when notes are shared on courses

**Security Features:**
- User isolation (notes belong to authenticated user)
- Transaction rollback on any failure for data consistency
- Input validation for required fields
- Safe JSON serialization with ToMap() method

**Status Codes:**
- 200 OK: Note created successfully
- 400 Bad Request: Invalid JSON, missing required fields
- 401 Unauthorized: Invalid or missing JWT token
- 409 Conflict: Note creation failure (constraint violations)
- 500 Internal Server Error: Database operations, AI generation, or processing failures

**AI Features:**
- Automatic content generation when both content and keywords are empty
- Context-aware generation using course, title, and subject information
- Fallback to manual content if AI generation fails

---

### POST `/unipilot/api/v1/notes/update`
**Purpose:** Updates a specific field of an existing note with ownership validation and audit tracking.

**Description:** Performs targeted field updates using raw SQL within a database transaction for data consistency. Only the note owner can update their notes with automatic timestamp tracking for audit trails.

**Request Body:**
- `id`: Note ID to update (string, required, converted to int)
- `column`: Database column name to update (string, required)
- `value`: New value for the specified column (string, required)

**Response:** 200 OK with no response body (success indicated by status code)

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Validates note ownership by user_id before allowing updates
- Updates specified column in `notes` table using raw SQL
- Automatically updates `updated_at` timestamp for audit trail
- Uses database transaction for atomicity and consistency

**Security Features:**
- Note ownership verification before allowing any updates
- Transaction rollback on any failure for data consistency
- Input validation for note ID format and conversion
- User isolation (users can only update their own notes)

**Security Warning:** Uses string interpolation for column name - column should be validated to prevent SQL injection attacks in production environments.

**Status Codes:**
- 200 OK: Note updated successfully
- 400 Bad Request: Invalid JSON body, malformed note ID
- 401 Unauthorized: Invalid or missing JWT token
- 500 Internal Server Error: Database operations failure, note not found

**Side Effects:**
- Modifies note record in database with timestamp update
- Logs successful updates with change details for audit trail
- No notification system integration (updates are silent)

---

## Social Features (Follow System)

### POST `/unipilot/api/v1/follow`
**Purpose:** Manages follow and unfollow operations between users with social engagement features.

**Description:** Toggles follow relationships, updates user statistics, calculates shared courses, and prepares notifications for social engagement. Prevents self-following and handles duplicate follow attempts gracefully with comprehensive social context.

**Request Body:**
- `followed_id`: ID of the user to follow/unfollow (uint, required)

**Response:**
- `success`: Boolean indicating operation success
- `message`: Descriptive message ("Followed successfully" or "Unfollowed successfully")

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Database Operations:**
- Checks existing follow relationship status
- Creates or removes follow record in database
- Updates follow statistics for both users (follower/following counts)
- Calculates shared courses between users for notification context

**Social Features:**
- Prevents users from following themselves
- Tracks shared courses for enhanced social context
- Maintains bidirectional follow statistics
- Prepares rich notifications with course sharing information

**Notification System (Prepared):**
- Code prepared for SSE notifications on new follows
- Includes shared course count in notification message
- Currently disabled pending Docker deployment

**Security Features:**
- Input validation for followed_id parameter
- Self-follow prevention for social integrity
- User authentication required for all operations
- Graceful handling of duplicate follow attempts

**Status Codes:**
- 200 OK: Follow/unfollow operation successful
- 400 Bad Request: Invalid JSON, missing followed_id, or self-follow attempt
- 401 Unauthorized: Invalid or missing JWT token
- 405 Method Not Allowed: Non-POST request
- 500 Internal Server Error: Database operations failure

**Social Analytics:**
- Logs follow actions for social engagement metrics
- Tracks shared course relationships for recommendation systems
- Monitors follow patterns for user discovery features

---

### GET `/unipilot/api/v1/followers`
**Purpose:** Retrieves a paginated list of users who follow the specified user with Redis caching.

**Description:** Implements Redis caching with database fallback for optimal performance on social feeds. Supports pagination through limit/offset parameters for handling large follower lists efficiently.

**Query Parameters:**
- `user_id`: ID of the user whose followers to retrieve (string, required)
- `limit`: Maximum number of followers to return (string, optional, default: 20)
- `offset`: Number of followers to skip for pagination (string, optional, default: 0)

**Response:**
- `followers`: Array of user objects representing the followers
- `total`: Total number of followers (for pagination calculations)

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Caching Strategy:**
- Redis key pattern: `followers:{user_id}`
- Cache structure: Hash with follower ID as field, JSON user object as value
- TTL: 10 minutes for social feed freshness
- Cache hit: Returns cached followers directly
- Cache miss: Queries database, populates cache, returns fresh data

**Performance Features:**
- Redis caching reduces database load for popular users
- Individual follower caching allows partial cache updates
- Non-blocking cache operations (warnings logged on failure)
- Pagination prevents memory issues with large follower lists

**Status Codes:**
- 200 OK: Followers retrieved successfully
- 400 Bad Request: Missing user_id parameter or invalid format
- 401 Unauthorized: Invalid or missing JWT token
- 405 Method Not Allowed: Non-GET request
- 500 Internal Server Error: Database query or Redis failures

---

### GET `/unipilot/api/v1/following`
**Purpose:** Retrieves a paginated list of users that the specified user follows with Redis caching.

**Description:** Implements Redis caching with database fallback for optimal performance on social feeds. Supports pagination through limit/offset parameters for handling large following lists efficiently.

**Query Parameters:**
- `user_id`: ID of the user whose following list to retrieve (string, required)
- `limit`: Maximum number of following users to return (string, optional, default: 20)
- `offset`: Number of following users to skip for pagination (string, optional, default: 0)

**Response:**
- `following`: Array of user objects representing users being followed
- `total`: Total number of users being followed (for pagination calculations)

**Authentication:** Required (AuthMiddleware) - extracts user from JWT token

**Caching Strategy:**
- Redis key pattern: `following:{user_id}`
- Cache structure: Hash with followed user ID as field, JSON user object as value
- TTL: 10 minutes for social feed freshness
- Cache hit: Returns cached following list directly
- Cache miss: Queries database, populates cache, returns fresh data

**Status Codes:**
- 200 OK: Following list retrieved successfully
- 400 Bad Request: Missing user_id parameter or invalid format
- 401 Unauthorized: Invalid or missing JWT token
- 405 Method Not Allowed: Non-GET request
- 500 Internal Server Error: Database query or Redis failures

---

### GET `/unipilot/api/v1/follow/status`
**Purpose:** Retrieves follow relationship status and statistics between users for UI state management.

**Description:** Provides comprehensive information for displaying follow buttons and user profile statistics. Returns follow relationship status along with follower/following counts for the target user with real-time accuracy.

**Query Parameters:**
- `user_id`: ID of the target user to check follow status against (string, required)

**Response:**
- `is_following`: Whether current user follows the target user (boolean)
- `followers_count`: Number of users following the target user (int)
- `following_count`: Number of users the target user follows (int)

**Authentication:** Required (AuthMiddleware) - extracts current user from JWT token

**Database Operations:**
- Checks follow relationship between current user and target user
- Retrieves follower count for target user
- Retrieves following count for target user
- No caching implemented (real-time status for UI accuracy)

**Use Cases:**
- Displaying follow/unfollow buttons in user interfaces
- Showing user profile statistics (follower/following counts)
- Social feed relationship indicators
- User discovery and recommendation systems

**Status Codes:**
- 200 OK: Follow status retrieved successfully
- 400 Bad Request: Missing user_id parameter or invalid format
- 401 Unauthorized: Invalid or missing JWT token
- 405 Method Not Allowed: Non-GET request
- 500 Internal Server Error: Database query failures

**Performance Notes:**
- No caching ensures real-time accuracy for UI state
- Optimized queries for fast response times
- Suitable for frequent status checks in social interfaces

---

## Course Management

### GET `/unipilot/api/v1/course/get`
**Purpose:** Retrieves all courses belonging to the authenticated user.

**Description:** Returns a list of all courses associated with the current user's account. Courses are converted to map format for consistent JSON serialization.

**Response:**
- `message`: Success message
- `courses`: Array of course objects (as maps)

**Authentication:** Required (AuthMiddleware)

**Database Operations:**
- Reads from `courses` table filtered by `user_id`

---

### POST `/unipilot/api/v1/course`
**Purpose:** Creates a new course for the authenticated user.

**Description:** Validates required fields, parses dates and numeric values, and stores the course in the database using a transaction for atomicity.

**Request Body:**
- `local_id`: Course local identifier (string, required)
- `name`: Course name (string)
- `code`: Course code (string, required)
- `color`: Course color (string)
- `semester`: Semester identifier (string, required)
- `schedule`: Course schedule (string)
- `credits`: Number of credits (string, converted to int)
- `location`: Course location (string)
- `start_date`: Start date in YYYY-MM-DD format (string, required)
- `end_date`: End date in YYYY-MM-DD format (string, required)
- `instructor`: Instructor name (string, required)
- `instructor_email`: Instructor email (string)

**Response:**
- `message`: Success message
- `course`: Created course object (as map)

**Authentication:** Required (AuthMiddleware)

**Database Operations:**
- Creates record in `courses` table
- Uses database transaction for atomicity

**Status Codes:**
- 200 OK: Course created successfully
- 400 Bad Request: Validation fails or date/numeric parsing fails
- 409 Conflict: Course creation fails (e.g., duplicate constraint)
- 500 Internal Server Error: Database operations fail

---

### POST `/unipilot/api/v1/course/update`
**Purpose:** Updates a specific field of a course.

**Description:** Updates a single column value for a course. Uses a database transaction and executes a raw SQL UPDATE statement.

**Request Body:**
- `id`: Course ID to update (string, converted to int)
- `column`: Database column name to update (string)
- `value`: New value for the column (string)

**Response:**
- 200 OK: Success (no body, logged)

**Authentication:** Required (AuthMiddleware)

**Database Operations:**
- Updates `courses` table using raw SQL UPDATE
- Uses database transaction for atomicity

**Status Codes:**
- 200 OK: Update successful
- 400 Bad Request: Invalid request body or course ID conversion fails
- 500 Internal Server Error: Database operations fail

**Security Note:** Uses string interpolation for column name - column should be validated to prevent SQL injection.

---

### POST `/unipilot/api/v1/course/link/request`
**Purpose:** Initiates a course sharing request by sending notifications to specified users.

**Description:** Generates or retrieves a link UUID for the course and sends course data via SSE notifications to the specified recipients.

**Request Body:**
- `course_code`: Code of the course to share (string, required)
- `users_id`: Array of user IDs to send the link request to ([]uint, required)

**Response:**
- `message`: Success message
- `course_id`: ID of the course being shared
- `link_id`: UUID link identifier for the course
- `recipients`: Array of recipient user IDs

**Authentication:** Required (AuthMiddleware)

**Database Operations:**
- Reads course by code from `courses` table
- Updates course `LinkID` if not already set

**Notifications:** Sends SSE notifications to all specified recipients via gRPC

**Status Codes:**
- 200 OK: Link request processed successfully
- 400 Bad Request: Invalid request body or course lookup fails

---

### POST `/unipilot/api/v1/course/link/accept`
**Purpose:** Accepts a course link request and synchronizes course assignments with documents.

**Description:** When a user accepts a course sharing link, this endpoint retrieves all assignments and their associated documents from the original course owner's course. The assignments are returned to the client for local synchronization. The original course owner is notified via SSE that their course link was accepted.

**Request Body:**
- `course.Course` object containing course metadata (user_id, code, etc.)

**Response:**
- `assignments`: Array of assignment objects with embedded documents

**Authentication:** Required (AuthMiddleware)

**Database Operations:**
- Reads assignments from `assignments` table filtered by `user_id` and `course_code`
- Reads documents associated with each assignment

**Notifications:** Sends SSE notification to original course owner via gRPC

**Status Codes:**
- 200 OK: Assignments retrieved successfully
- 400 Bad Request: Invalid request body or database queries fail

---

## AI Services (Node.js Microservice)

### GET `/health`
**Purpose:** Health check endpoint for AI service monitoring and load balancer health checks.

**Description:** Returns service status and identification for operational monitoring. Used by orchestration systems to verify service availability and readiness.

**Request Body:** None required

**Response:**
- `status`: Service health status ("healthy")
- `service`: Service identifier ("ai-chat")

**Authentication:** None required (public health check endpoint)

**Status Codes:**
- 200 OK: Service is healthy and operational

---

### POST `/unipilot/ai/v1`
**Purpose:** Main AI chat endpoint providing streaming conversational AI with RAG capabilities.

**Description:** Integrates Google Gemini AI with Qdrant vector database for context-aware responses. Features streaming text generation for real-time responses, RAG integration for assignment-specific knowledge retrieval, and tool-based information access from vector database with assignment context injection.

**Request Body:**
- `messages`: Conversation history in AI SDK format (Array, required)
- `assignment`: Assignment context object (Object, required)
- `assignment.RemoteID`: Assignment ID for RAG collection targeting (string, required)
- `assignment.Title`: Assignment title for context (string, required)
- `assignment.Course`: Course information object (Object, required)
- `assignment.Course.Name`: Course name (string, required)
- `assignment.Course.Code`: Course code (string, required)
- `assignment.Type`: Assignment type information (Object, required)
- `assignment.Type.Name`: Assignment type name (string, required)
- `assignment.Priority`: Assignment priority level (string, required)
- `assignment.Deadline`: Assignment due date (string, required)
- `assignment.Todo`: Assignment description/tasks (string, required)
- `assignment.StatusName`: Current assignment status (string, required)

**Response:** Streaming AI response with tool integration (Server-Sent Events format)

**Authentication:** Handled by calling Go backend service (no direct authentication)

**AI Processing Pipeline:**
1. **Message Conversion:** Converts conversation history to AI SDK format
2. **Context Building:** Creates assignment-specific system prompt for personalized assistance
3. **RAG Integration:** Configures tool for knowledge base access using Qdrant vector search
4. **Streaming Generation:** Uses Google Gemini 2.0 Flash Lite for real-time AI responses
5. **Tool Execution:** Automatically searches assignment documents when information is needed

**RAG Tool Integration:**
- **Tool Name:** `getInformation`
- **Function:** Searches assignment-specific Qdrant collection for relevant document chunks
- **Input:** User's question extracted from conversation context
- **Output:** Contextual information from document embeddings for AI response generation

**AI Configuration:**
- **Model:** Google Gemini 2.0 Flash Lite for fast, high-quality responses
- **Max Tokens:** 4000 tokens for comprehensive responses
- **Temperature:** 0.7 for balanced creativity and accuracy
- **System Prompt:** Assignment-specific context including course, type, priority, and deadline information

**Status Codes:**
- 200 OK: Streaming AI response initiated successfully
- 500 Internal Server Error: AI service unavailable, processing failures, or tool execution errors

**Performance Features:**
- **Streaming Responses:** Real-time text generation for immediate user feedback
- **Context Awareness:** Assignment-specific prompts for relevant academic assistance
- **Tool Integration:** Automatic document search when additional information is needed
- **Error Resilience:** Graceful handling of AI service failures with detailed error messages

**CORS Configuration:**
- **Origins:** Wildcard (*) for development/testing environments
- **Methods:** GET, POST, OPTIONS for comprehensive web client support
- **Headers:** Content-Type, Authorization, User-Agent, x-session-id for secure communication
- **Credentials:** Enabled for authenticated cross-origin requests
