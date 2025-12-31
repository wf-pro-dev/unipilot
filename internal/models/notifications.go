package models

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationLink   NotificationType = "link"
	NotificationUnlink NotificationType = "unlink"

	NotificationCourse NotificationType = "course"

	NotificationAssignment       NotificationType = "assignment"
	NotificationAssignmentUpdate NotificationType = "assignment_update"

	NotificationNote       NotificationType = "note"
	NotificationNoteUpdate NotificationType = "note_update"

	NotificationDocument       NotificationType = "document"
	NotificationDocumentUpdate NotificationType = "document_update"

	NotificationFollow NotificationType = "follow"

	NotificationSync NotificationType = "sync"
)

type LocalNotification struct {
	gorm.Model
	SenderID  uint             `gorm:"not null" json:"sender_id"`
	Entity    Entity           `gorm:"not null" json:"entity"`
	EntityID  uint             `gorm:"not null" json:"entity_id"`
	Type      NotificationType `gorm:"not null" json:"type"`
	Title     string           `gorm:"not null" json:"title"`
	Action    string           `gorm:"not null" json:"action"` // create, update
	Message   string           `gorm:"not null" json:"message"`
	Read      bool             `gorm:"default:false" json:"read"`
	Data      string           `json:"data"`
	Sender    User             `gorm:"foreignKey:SenderID;references:ID" json:"sender"`
	ExpiresAt *time.Time       `gorm:"index"`
}

func (n *LocalNotification) ToMap() map[string]string {
	return map[string]string{
		"sender_id":  strconv.Itoa(int(n.SenderID)),
		"entity":     string(n.Entity),
		"entity_id":  strconv.Itoa(int(n.EntityID)),
		"type":       string(n.Type),
		"title":      n.Title,
		"action":     n.Action,
		"message":    n.Message,
		"expires_at": n.ExpiresAt.Format(time.RFC3339),
		"created_at": n.CreatedAt.Format(time.RFC3339),
		"updated_at": n.UpdatedAt.Format(time.RFC3339),
	}
}
