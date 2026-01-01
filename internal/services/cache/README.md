# Cache Service

Centralized Redis caching service organized by resource with strict nomenclature.

## Nomenclature

### Function Names
Format: `<method><resource_1><resource_2>` (camelCase)
- Examples: `GetUserFollowing`, `SetAssignmentsCourse`, `GetNotesCourse`

### Cache Keys
Format: `resource1:resource2:...:id` (following model hierarchy)
- Examples: `user:following:%d`, `course:assignments:%d`, `course:notes:%d`

## Architecture

- **Centralized Cache**: Single `Cache` struct holds Redis client
- **Resource-Based Files**: Each resource has its own file
- **Separation of Concerns**: Each resource is independent and reusable
- **No DB Logic**: Cache service only handles Redis operations

## File Structure

- `cache.go` - Centralized Cache struct
- `users.go` - Users resource
- `followers.go` - User followers resource
- `following.go` - User following resource
- `courses.go` - Courses resource
- `assignments.go` - Course assignments resource
- `notes.go` - Course notes resource
- `linked_courses.go` - User linked courses resource
- `keys.go` - Cache key patterns
- `ttl.go` - TTL constants

## Usage

### Initialization

```go
import "unipilot/internal/services/cache"

// Initialize centralized cache
cacheService := cache.New(redisClient)
```

### Users Resource

```go
// Get all users from cache
usersHash, err := cacheService.GetUsers(ctx)

// Set a user in cache
err := cacheService.SetUsers(ctx, userID, user)

// Delete a user from cache
err := cacheService.DeleteUsers(ctx, userID)

// Set expiration
err := cacheService.SetExpirationUsers(ctx)
```

### User Followers Resource

```go
// Get followers for a user
followersHash, err := cacheService.GetUserFollowers(ctx, userID)

// Set a follower in cache
err := cacheService.SetUserFollowers(ctx, userID, followerID, follower)

// Delete a follower from cache
err := cacheService.DeleteUserFollowers(ctx, userID, followerID)

// Set expiration
err := cacheService.SetExpirationUserFollowers(ctx, userID)
```

### User Following Resource

```go
// Get following list for a user
followingHash, err := cacheService.GetUserFollowing(ctx, userID)

// Set a followed user in cache
err := cacheService.SetUserFollowing(ctx, userID, followedID, followed)

// Delete a followed user from cache
err := cacheService.DeleteUserFollowing(ctx, userID, followedID)

// Set expiration
err := cacheService.SetExpirationUserFollowing(ctx, userID)
```

### Courses Resource

```go
// Get a course from cache
course, err := cacheService.GetCourse(ctx, courseID)

// Set a course in cache
err := cacheService.SetCourse(ctx, courseID, course)

// Delete a course from cache
err := cacheService.DeleteCourse(ctx, courseID)

// Get linked course IDs for a course
linkedIDs, err := cacheService.GetCourseLinkedCourseIDs(ctx, courseID)

// Set linked course IDs for a course
err := cacheService.SetCourseLinkedCourseIDs(ctx, courseID, linkedIDs)

// Delete linked course IDs
err := cacheService.DeleteCourseLinkedCourseIDs(ctx, courseID)
```

### Course Assignments Resource

```go
// Get assignments for a course
assignments, err := cacheService.GetAssignmentsCourse(ctx, courseID)

// Set assignments for a course
err := cacheService.SetAssignmentsCourse(ctx, courseID, assignments)

// Delete assignments cache for a course
err := cacheService.DeleteAssignmentsCourse(ctx, courseID)
```

### Course Notes Resource

```go
// Get notes for a course
notes, err := cacheService.GetNotesCourse(ctx, courseID)

// Set notes for a course
err := cacheService.SetNotesCourse(ctx, courseID, notes)

// Delete notes cache for a course
err := cacheService.DeleteNotesCourse(ctx, courseID)
```

### User Linked Courses Resource

```go
// Get user's linked course IDs
courseIDs, err := cacheService.GetUserLinkedCourses(ctx, userID)

// Set user's linked course IDs
err := cacheService.SetUserLinkedCourses(ctx, userID, courseIDs)

// Delete user's linked course IDs
err := cacheService.DeleteUserLinkedCourses(ctx, userID)

// Delete all course data (course, linked IDs, assignments, notes)
err := cacheService.DeleteAllCourseData(ctx, courseID)
```

## Cache Keys

All cache keys follow model hierarchy: `resource1:resource2:...:id`

- `users` - Users hash
- `user:followers:{userID}` - User's followers hash
- `user:following:{userID}` - User's following hash
- `user:linked_courses:{userID}` - User's linked course IDs
- `course:{courseID}` - Course data
- `course:linked:{courseID}` - Course's linked course IDs
- `course:assignments:{courseID}` - Course's assignments
- `course:notes:{courseID}` - Course's notes

## TTL Values

All TTL values are defined in `ttl.go`:
- Users: 2 hours
- User Followers/Following: 30 minutes
- User Linked Courses: 15 minutes
- Course: 30 minutes
- Course Linked Course IDs: 30 minutes
- Course Assignments: 20 minutes
- Course Notes: 20 minutes

## Design Principles

1. **Strict Nomenclature**: Function names and cache keys follow consistent patterns
2. **Model Hierarchy**: Cache keys reflect model relationships (user:followers, course:assignments)
3. **Resource Separation**: Each resource has its own file, reusable across workflows
4. **No DB Logic**: Cache services only handle Redis operations
5. **Clear Naming**: Function names clearly indicate operation and resources involved
