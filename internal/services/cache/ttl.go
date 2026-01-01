package cache

import "time"

// TTL (Time To Live) constants organized by resource.
const (
	// LinkID -> Users (CRITICAL: write-maintained, long TTL)
	TTLLinkUsers = 24 * time.Hour

	// Users resource (MEDIUM: read-heavy, moderate TTL)
	TTLUsers = 2 * time.Hour

	// User -> Followers (HIGH: read-heavy social graph)
	TTLUserFollowers = 30 * time.Minute

	// User -> Following (HIGH: read-heavy social graph)
	TTLUserFollowing = 30 * time.Minute

	// User -> LinkedCourses (MEDIUM: read-heavy complex query)
	TTLUserLinkedCourses = 15 * time.Minute
)
