# Database Schema & Caching Strategy

## Database Tables

### Users Table
The `users` table stores user account information and authentication data.

**Primary Key:** `id` (auto-increment)

**Key Fields:**
- `id`: Unique user identifier
- `username`: Unique username for authentication
- `email`: User's email address
- `password`: Bcrypt-hashed password
- `university`: User's university affiliation
- `language`: User's preferred language
- `created_at`: Account creation timestamp
- `updated_at`: Last profile update timestamp

**Relationships:**
- One-to-many with `courses` table (`user_id` foreign key)

### Courses Table
The `courses` table stores course information associated with users.

**Primary Key:** `id` (auto-increment)

**Key Fields:**
- `id`: Unique course identifier
- `user_id`: Foreign key to `users` table
- `local_id`: Course local identifier
- `name`: Course name
- `code`: Course code (used for sharing and identification)
- `color`: UI color for course display
- `semester`: Semester identifier
- `schedule`: Course schedule information
- `credits`: Number of course credits
- `location`: Course location/classroom
- `start_date`: Course start date
- `end_date`: Course end date
- `instructor`: Instructor name
- `instructor_email`: Instructor email address
- `link_id`: UUID for course sharing (nullable)
- `created_at`: Course creation timestamp
- `updated_at`: Last course update timestamp

**Relationships:**
- Many-to-one with `users` table (`user_id` foreign key)
- One-to-many with `assignments` table (course_code relationship)

### Assignments Table
The `assignments` table stores assignment/task information associated with users and courses.

**Primary Key:** `id` (auto-increment)

**Key Fields:**
- `id`: Unique assignment identifier
- `user_id`: Foreign key to `users` table (assignment owner)
- `local_id`: Local identifier for client-side synchronization
- `title`: Assignment title/name
- `todo`: Assignment description and task details
- `deadline`: Assignment due date
- `course_code`: Associated course identifier (links to courses)
- `type_name`: Assignment category/type (e.g., "homework", "project")
- `status_name`: Current assignment status (e.g., "pending", "completed")
- `priority`: Assignment priority level
- `link`: Related URL or resource link
- `parent_id`: Foreign key to parent assignment (for sub-assignments)
- `created_at`: Assignment creation timestamp
- `updated_at`: Last assignment update timestamp

**Relationships:**
- Many-to-one with `users` table (`user_id` foreign key)
- Many-to-one with `courses` table (via `course_code`)
- Self-referential relationship (`parent_id` for hierarchical assignments)

**Special Features:**
- Hierarchical structure support through `parent_id`
- Course integration for assignment sharing and notifications
- Local ID system for client-side synchronization

### Documents Table
The `documents` table stores document/file metadata and cloud storage references.

**Primary Key:** `id` (auto-increment)

**Key Fields:**
- `id`: Unique document identifier
- `assignment_id`: Foreign key to `assignments` table (server assignment ID)
- `local_assignment_id`: Local assignment ID for client synchronization
- `local_id`: Local document identifier for client-side management
- `user_id`: Foreign key to `users` table (document owner)
- `type`: Document type/category (e.g., "pdf", "image", "text")
- `file_name`: Original file name as uploaded by user
- `file_type`: MIME type for proper file handling
- `file_size`: File size in bytes for storage tracking
- `version`: Document version number for version control
- `is_original`: Boolean flag indicating if this is the original document
- `file_path`: Local file path (used for temporary processing)
- `storage_key`: AWS S3 storage key for cloud file location
- `created_at`: Document upload timestamp
- `updated_at`: Last document update timestamp

**Relationships:**
- Many-to-one with `users` table (`user_id` foreign key)
- Many-to-one with `assignments` table (`assignment_id` foreign key)

**Cloud Storage Integration:**
- Files stored in AWS S3 with organized directory structure
- Storage path pattern: `users_data/user_{id}/documents/assign_{id}/{timestamp}_{filename}`
- Unique timestamped filenames prevent conflicts and overwrites
- Storage keys used for secure file access without exposing file paths

**RAG (Retrieval-Augmented Generation) Integration:**
- Documents processed into AI embeddings for semantic search
- Vector storage in Qdrant database with assignment-based collections
- Collection naming: `unipilot-qdrant-db-{assignment_id}`
- 768-dimensional vectors with cosine distance similarity
- Enables AI-powered document search and retrieval

**Qdrant Vector Database Configuration:**
- **Client Connection:** gRPC-based client using official Qdrant Go library
- **Host Configuration:** Retrieved from `QDRANT_HOST` environment variable
- **Port:** Standard gRPC port 6334 for vector operations
- **Connection Management:** Singleton client pattern for efficient resource usage
- **Error Handling:** Comprehensive error handling for connection failures and missing configuration
- **Use Cases:** Document embeddings, semantic search, and RAG-powered content retrieval

### Notes Table
The `notes` table stores user-generated notes with AI-powered content generation capabilities.

**Primary Key:** `id` (auto-increment)

**Key Fields:**
- `id`: Unique note identifier
- `user_id`: Foreign key to `users` table (note owner)
- `course_code`: Associated course identifier for organization
- `title`: Note title/heading
- `subject`: Note subject or topic area
- `content`: Main note content (can be AI-generated)
- `keywords`: Note keywords for search and categorization (can be AI-generated)
- `videos`: Associated video links or references
- `created_at`: Note creation timestamp
- `updated_at`: Last note update timestamp

**Relationships:**
- Many-to-one with `users` table (`user_id` foreign key)
- Associated with courses via `course_code` (soft relationship)

**AI Integration (Google Gemini):**
- Automatic content generation when both content and keywords are empty
- Uses title, subject, and course code as context for AI generation
- Provides intelligent note scaffolding for educational content
- Graceful fallback to manual content if AI generation fails

**Content Generation Process:**
1. User provides title, subject, and course code
2. If content and keywords are empty, system calls Google Gemini API
3. AI generates relevant educational content based on provided context
4. Generated content and keywords are stored alongside user-provided metadata
5. Manual content override available for user customization

**Notification System (Prepared):**
- Infrastructure prepared for real-time note sharing notifications
- Course-based notification distribution to linked users
- Currently disabled pending Docker deployment integration

### Follow Relationships (Social System)
The follow system implements user-to-user relationships for social features and collaboration.

**Database Structure:**
- **Primary Table:** `follows` (junction table for many-to-many relationships)
- **Key Fields:**
  - `follower_id`: Foreign key to `users` table (user who follows)
  - `followed_id`: Foreign key to `users` table (user being followed)
  - `created_at`: Follow relationship creation timestamp

**User Statistics Integration:**
- **Follower Count:** Cached count of users following a specific user
- **Following Count:** Cached count of users a specific user follows
- **Statistics Updates:** Automatic recalculation when follow relationships change
- **Performance Optimization:** Denormalized counts for fast profile display

**Social Features:**
- **Self-Follow Prevention:** Business logic prevents users from following themselves
- **Duplicate Prevention:** Database constraints prevent duplicate follow relationships
- **Bidirectional Tracking:** Maintains statistics for both follower and followed users
- **Course Integration:** Calculates shared courses between users for social context

**Relationships:**
- Many-to-many relationship between `users` table (self-referential)
- Supports complex social network analysis and user discovery
- Enables course sharing recommendations based on follow relationships

## Redis Caching Strategy

### Users Cache
**Key Pattern:** `users` (Redis Hash)
**Structure:** Hash with user ID as field, JSON user object as value
**TTL:** 1 hour
**Purpose:** Cache user listings for the GET /users endpoint

**Cache Operations:**
- **Read:** `HGETALL users` - retrieves all cached users
- **Write:** `HSET users {user_id} {json_data}` - caches individual user
- **Expiry:** `EXPIRE users 3600` - sets 1-hour TTL

**Cache Strategy:**
1. Cache hit: Return all users from Redis hash, exclude current user
2. Cache miss: Query database, populate cache, return results
3. Individual user caching allows partial cache updates
4. Non-blocking cache operations (failures logged as warnings)

**Data Format:**
```json
{
  "id": 1,
  "username": "john_doe",
  "email": "john@university.edu",
  "university": "Example University",
  "language": "en",
  "courses_code": ["CS101", "MATH201"],
  "created_at": "2023-01-15T10:30:00Z",
  "updated_at": "2023-01-20T14:45:00Z"
}
```

### User Profile Cache
**Key Pattern:** `user:{user_id}` (Redis String)
**Structure:** JSON-serialized user object
**TTL:** Varies by operation
**Purpose:** Cache individual user profiles for authentication and profile operations

### Follow Relationships Cache
**Key Patterns:** 
- `followers:{user_id}` (Redis Hash) - Users who follow the specified user
- `following:{user_id}` (Redis Hash) - Users that the specified user follows

**Structure:** Hash with user ID as field, JSON user object as value
**TTL:** 10 minutes for social feed freshness
**Purpose:** Cache follower/following lists for social feeds and user discovery

**Cache Operations:**
- **Read:** `HGETALL followers:{user_id}` or `HGETALL following:{user_id}`
- **Write:** `HSET followers:{user_id} {follower_id} {json_data}`
- **Expiry:** `EXPIRE followers:{user_id} 600` (10 minutes)

**Cache Strategy:**
1. Cache hit: Return cached social connections directly
2. Cache miss: Query database relationships, populate cache, return results
3. Individual user caching allows partial cache updates
4. Non-blocking cache operations (failures logged as warnings)

**Performance Benefits:**
- **Reduced Database Load:** Social feeds served from cache for popular users
- **Fast Social Discovery:** Sub-second response times for follow lists
- **Scalable Social Features:** Cache layer handles high social interaction volume

## Performance Considerations

### Database Queries
- **Users Listing:** Single query with JOIN to get course codes
- **User Authentication:** Indexed lookup by username
- **Course Operations:** Filtered by user_id for security

### Caching Benefits
- **Reduced Database Load:** Frequent user listings served from cache
- **Improved Response Time:** Redis access ~1ms vs database ~10-50ms
- **Scalability:** Cache layer handles high read traffic

### Cache Invalidation
- **User Updates:** Individual user cache refresh on profile changes
- **Course Changes:** User cache includes course codes, may need refresh
- **TTL Strategy:** 1-hour expiry balances freshness with performance

## Security Considerations

### Data Protection
- **Password Hashing:** bcrypt with secure defaults
- **Sensitive Field Exclusion:** Password hashes never cached or returned in API responses
- **User Isolation:** Users can only access their own data and public user listings

### Query Security
- **Parameterized Queries:** Prevents SQL injection for most operations
- **Input Validation:** Required for raw SQL operations (user/course updates)
- **Context-Based Authorization:** User context from JWT tokens ensures proper access control

## Monitoring & Maintenance

### Cache Health
- **Hit Rate Monitoring:** Track cache hit/miss ratios
- **Error Logging:** Non-blocking cache operations log warnings on failure
- **Memory Usage:** Monitor Redis memory consumption for user data

### Database Performance
- **Index Optimization:** Ensure proper indexing on user_id, username, course_code
- **Query Performance:** Monitor slow query logs for optimization opportunities
- **Connection Pooling:** Proper database connection management
