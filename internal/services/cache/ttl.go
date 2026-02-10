package cache

import "time"

// TTL (Time To Live) constants organized by resource.
const (

	// Users resource (MEDIUM: read-heavy, moderate TTL)
	TTLUser = 2 * time.Hour

	// User -> Friends (HIGH: read-heavy social graph)
	TTLUserFriends = 1 * time.Hour

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
