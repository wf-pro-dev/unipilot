package models

import (
	"log"
	"time"

	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// AssignmentType defines types like HW, Exam
type AssignmentType struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"unique;not null"`
	Color    string `gorm:"not null"`
	NotionID string
}

func Get_AssignmentType_byName(name string, db *gorm.DB) *AssignmentType {
	assignmentType := &AssignmentType{}
	err := db.Where("name = ?", name).First(assignmentType).Error
	if err != nil {
		log.Fatalln("Error getting assignment type with name: ", err)
		return nil
	}
	return assignmentType
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

func Get_AssignmentStatus_byName(name string, db *gorm.DB) *AssignmentStatus {
	assignmentStatus := &AssignmentStatus{}
	err := db.Where("name = ?", name).First(assignmentStatus).Error
	if err != nil {
		log.Fatalln("Error getting assignment status with name: ", err)
		return nil
	}
	return assignmentStatus
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

// SyncLog tracks changes for synchronization
type SyncLog struct {
	gorm.Model
	UserID        uint      `gorm:"not null"`
	User          user.User `gorm:"foreignKey:UserID;references:ID"`
	DeviceID      uint      `gorm:"not null"`
	Action        string    `gorm:"not null"` // create, update, delete
	TableName     string    `gorm:"not null"`
	RecordID      uint      `gorm:"not null"`
	SyncTimestamp time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Synced        bool      `gorm:"default:false"`
}
