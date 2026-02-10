package cache

import (
	"fmt"
)

// Cache key patterns following model hierarchy: resource1:resource2:...:id
const (
	KeyUser = "user:%s"

	KeyUserClusters = "user:%s:clusters"

	// User -> Followers (HIGH: read-heavy social graph)
	KeyUserFriends = "user:%s:friends"
)

const (

	// Course
	KeyCourse = "course:%s"

	KeyClusterCourses = "cluster:%s:courses" // Stores Set of Course ID

	KeyClusterUsers = "cluster:%s:users" // Stores Set of User ID

	// Course -> LinkedCourses (MEDIUM: read-heavy complex query)
	KeyCourseLinks = "course:%s:links"

	// Course -> assignments
	KeyCourseAssignments = "course:%s:assignments"

	// Course -> notes
	KeyCourseNotes = "course:%s:notes"
)

const (
	// Assignments
	KeyAssignment = "assignment:%s"
)

const (
	// Notes
	KeyNote = "note:%s"
)

const (
	// Progress cache
	KeyProgress = "progress:%s"
)

// FormatKey formats a cache key with the given identifier (string).
func FormatKey(pattern string, id string) string {
	return fmt.Sprintf(pattern, id)
}
