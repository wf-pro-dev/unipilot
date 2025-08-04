package course

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending" // Needs to be synced
	SyncStatusSynced  SyncStatus = "synced"  // Already synced
)

type LocalCourse struct {
	gorm.Model
	RemoteID        uint
	Code            string `gorm:"index:idx_course_code_active,unique,where:deleted_at IS NULL"`
	Name            string `gorm:"not null"`
	NotionID        string
	Duration        string
	RoomNumber      string
	Color           string
	StartDate       time.Time
	EndDate         time.Time
	Credits         int
	Schedule        string
	Semester        string
	Instructor      string
	InstructorEmail string
	SyncStatus      SyncStatus `gorm:"not null;default:'pending'"`
}

// BeforeCreate is a GORM hook that runs before creating a record
func (c *LocalCourse) BeforeCreate(tx *gorm.DB) error {
	// Check if a course with the same code exists (including soft-deleted ones)
	var existingCourse LocalCourse
	if err := tx.Unscoped().Where("code = ?", c.Code).First(&existingCourse).Error; err == nil {
		// A course with this code exists, check if it's soft-deleted
		if existingCourse.DeletedAt.Valid {
			// If it's soft-deleted, we can reuse the code
			return nil
		}
		// If it's not soft-deleted, return an error
		return fmt.Errorf("course with code '%s' already exists", c.Code)
	}
	return nil
}

func (c *LocalCourse) ToMap() map[string]string {
	return map[string]string{
		"remote_id":        strconv.Itoa(int(c.RemoteID)),
		"code":             c.Code,
		"name":             c.Name,
		"notion_id":        c.NotionID,
		"duration":         c.Duration,
		"room_number":      c.RoomNumber,
		"color":            c.Color,
		"start_date":       c.StartDate.Format(time.DateOnly),
		"end_date":         c.EndDate.Format(time.DateOnly),
		"schedule":         c.Schedule,
		"credits":          strconv.Itoa(int(c.Credits)),
		"semester":         c.Semester,
		"instructor":       c.Instructor,
		"instructor_email": c.InstructorEmail,
		"sync_status":      string(c.SyncStatus),
	}
}
