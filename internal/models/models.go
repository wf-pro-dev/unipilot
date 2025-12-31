package models

import (
	"time"

	"gorm.io/gorm"
)

type Entity string

const (
	EntityUser         Entity = "user"
	EntityAssignment   Entity = "assignment"
	EntityCourse       Entity = "course"
	EntityNote         Entity = "note"
	EntityDocument     Entity = "document"
	EntityNotification Entity = "notification"
)

// Device tracks sync status for different devices
type Device struct {
	gorm.Model
	UserID     uint   `gorm:"not null"`
	User       User   `gorm:"foreignKey:UserID;references:ID"`
	DeviceID   string `gorm:"unique;not null"`
	DeviceName string
	LastSync   *time.Time
	SyncToken  string
}
