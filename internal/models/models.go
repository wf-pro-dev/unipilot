package models

import (
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/user"
)

// AssignmentType defines types like HW, Exam
type AssignmentType struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Color    string `gorm:"not null"`
	NotionID string
}

func (a *AssignmentType) ToMap() map[string]string {
	return map[string]string{
		"id":    a.NotionID,
		"name":  a.Name,
		"color": a.Color,
	}
}

// AssignmentStatus defines statuses like Not Started, In Progress
type AssignmentStatus struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Color    string `gorm:"not null"`
	NotionID string
}

func (a *AssignmentStatus) ToMap() map[string]string {
	return map[string]string{
		"id":    a.NotionID,
		"name":  a.Name,
		"color": a.Color,
	}
}

// Device tracks sync status for different devices
type Device struct {
	gorm.Model
	UserID     uint      `gorm:"not null"`
	User       user.User `gorm:"foreignKey:UserID;references:ID"`
	DeviceID   string    `gorm:"unique;not null"`
	DeviceName string
	LastSync   *time.Time
	SyncToken  string
}
