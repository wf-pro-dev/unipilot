package models

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

// AssignmentType defines types like HW, Exam
type LocalAssignmentType struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Color    string `gorm:"not null"`
	NotionID string
}

func (a *LocalAssignmentType) ToMap() map[string]string {
	return map[string]string{
		"id":    a.NotionID,
		"name":  a.Name,
		"color": a.Color,
	}
}

// AssignmentStatus defines statuses like Not Started, In Progress
type LocalAssignmentStatus struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Color    string `gorm:"not null"`
	NotionID string
}

func (a *LocalAssignmentStatus) ToMap() map[string]string {
	return map[string]string{
		"id":    a.NotionID,
		"name":  a.Name,
		"color": a.Color,
	}
}

type Entity string

const (
	EntityAssignment Entity = "assignment"
	EntityCourse     Entity = "course"
	EntityDocument   Entity = "document"
)

type LocalUpdate struct {
	gorm.Model
	Entity   Entity
	EntityID uint   `gorm:"not null"`
	Action   string `gorm:"not null"` // create, update, delete
	Column   string
	Value    string
	SyncedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Synced   bool      `gorm:"default:false"`
}

func (u *LocalUpdate) ToMap() map[string]string {
	return map[string]string{
		"entity":    string(u.Entity),
		"entity_id": strconv.Itoa(int(u.EntityID)),
		"action":    u.Action,
		"column":    u.Column,
		"value":     u.Value,
		"synced":    strconv.FormatBool(u.Synced),
	}
}
