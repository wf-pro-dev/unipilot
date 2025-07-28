package course

import (
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
	RemoteID        uint   `gorm:"unique"` // Empty until synced
	Code            string `gorm:"unique"`
	Name            string `gorm:"not null"`
	NotionID        string `gorm:"unique"` // Empty until synced
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
