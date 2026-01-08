package cache

import "time"

// TTL (Time To Live) constants organized by resource.
const (

	// Users resource (MEDIUM: read-heavy, moderate TTL)
	TTLUsers = 2 * time.Hour

	// User -> Followers (HIGH: read-heavy social graph)
	TTLUserFollowers = 30 * time.Minute

	// User -> Following (HIGH: read-heavy social graph)
	TTLUserFollowing = 30 * time.Minute

	// User -> LinkedCourses (MEDIUM: read-heavy complex query)
	TTLUserCoursesLinked = 1 * time.Hour

	// TTL for progress cache
	TTLProgress = 30 * time.Minute
)

const (

	// Course
	TTLCourse = 30 * time.Minute

	// Cluster (course) -> Users
	TTLClusterUsers = 1 * time.Hour

	// Course -> LinkedCourses
	TTLCourseLinks = 30 * time.Minute

	// Course -> assignments
	TTLCourseAssignments = 15 * time.Minute

	// Course -> notes
	TTLCourseNotes = 15 * time.Minute
)

const (

	// Assignments
	TTLAssignment = 5 * time.Minute
)

const (

	// Notes
	TTLNote = 5 * time.Minute
)
