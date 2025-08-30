package models

import (
	"strconv"
	"time"

	"unipilot/internal/models/user"

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
	EntityNote       Entity = "note"
	EntityUser       Entity = "user"
	EntityFollow     Entity = "follow"
)

type LocalUpdate struct {
	gorm.Model
	Entity      Entity
	EntityID    uint   `gorm:"not null"`
	Action      string `gorm:"not null"` // create, update, delete
	Column      string
	Value       string
	RetryCount  int `gorm:"default:0"`
	NextRetryAt time.Time
	LastError   string
}

func (u *LocalUpdate) ToMap() map[string]string {
	return map[string]string{
		"entity":        string(u.Entity),
		"entity_id":     strconv.Itoa(int(u.EntityID)),
		"action":        u.Action,
		"column":        u.Column,
		"value":         u.Value,
		"retry_count":   strconv.Itoa(u.RetryCount),
		"next_retry_at": u.NextRetryAt.Format(time.RFC3339),
		"last_error":    u.LastError,
	}
}

type LocalNotification struct {
	gorm.Model
	SenderID uint      `gorm:"not null" json:"sender_id"`
	Entity   Entity    `gorm:"not null" json:"entity"`
	EntityID uint      `gorm:"not null" json:"entity_id"`
	Title    string    `gorm:"not null" json:"title"`
	Subtitle string    `json:"subtitle"`
	Action   string    `gorm:"not null" json:"action"` // create, update
	Message  string    `gorm:"not null" json:"message"`
	Sender   user.User `gorm:"foreignKey:SenderID;references:ID" json:"sender"`
}

func (n *LocalNotification) ToMap() map[string]string {
	return map[string]string{
		"sender_id":  strconv.Itoa(int(n.SenderID)),
		"entity":     string(n.Entity),
		"entity_id":  strconv.Itoa(int(n.EntityID)),
		"title":      n.Title,
		"action":     n.Action,
		"message":    n.Message,
		"created_at": n.CreatedAt.Format(time.RFC3339),
		"updated_at": n.UpdatedAt.Format(time.RFC3339),
	}
}

