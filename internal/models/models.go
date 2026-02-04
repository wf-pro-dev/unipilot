package models

import (
	"time"
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
	Base
	UserID     string `gorm:"not null"`
	DeviceID   string `gorm:"unique;not null"`
	DeviceName string
	LastSync   *time.Time
	SyncToken  string

	User *User `gorm:"foreignKey:UserID;references:ID" validate:"-"`
}
