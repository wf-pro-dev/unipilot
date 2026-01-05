package cache

import "fmt"

// Cache key patterns following model hierarchy: resource1:resource2:...:id
const (
	// Course -> Users mapping (CRITICAL: write-heavy shared data)
	KeyCoursesLinkedUsers = "courses_linked:%s:users"

	// Users resource (MEDIUM: read-heavy shared data)
	KeyUsers = "users"

	// User -> Followers (HIGH: read-heavy social graph)
	KeyUserFollowers = "user:followers:%s"

	// User -> Following (HIGH: read-heavy social graph)
	KeyUserFollowing = "user:following:%s"

	// User -> LinkedCourses (MEDIUM: read-heavy complex query)
	KeyCoursesLinked = "courses_linked:%s"

	// Progress cache
	KeyProgress = "progress:%d"
)

// FormatKey formats a cache key with the given identifier (string).
func FormatKey(pattern string, id string) string {
	return fmt.Sprintf(pattern, id)
}
