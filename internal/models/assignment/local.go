package assignment

import (
	"strconv"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"

	"gorm.io/gorm"
)

type LocalAssignment struct {
	gorm.Model
	RemoteID   uint   `gorm:"unique"`
	Title      string `gorm:"not null"`
	Todo       string
	Deadline   time.Time `gorm:"not null;index"`
	Link       string    `gorm:"default:https://acconline.austincc.edu/ultra/stream"`
	CourseCode string    `gorm:"not null;index"`
	TypeName   string    `gorm:"not null"`
	StatusName string    `gorm:"not null"`
	Priority   string    `gorm:"default:medium"`
	ParentID   uint      `gorm:"default:0"`

	Course    course.LocalCourse           `gorm:"foreignKey:CourseCode;references:Code"`
	Type      models.LocalAssignmentType   `gorm:"foreignKey:TypeName;references:Name"`
	Status    models.LocalAssignmentStatus `gorm:"foreignKey:StatusName;references:Name"`
	Documents []document.LocalDocument     `gorm:"foreignKey:AssignmentID;references:ID"`
	Parent    *LocalAssignment             `gorm:"foreignKey:ParentID;references:ID"`
	Children  []*LocalAssignment           `gorm:"foreignKey:ParentID;references:ID"`
}

func (a *LocalAssignment) ToMap() map[string]string {
	return map[string]string{
		"id":          strconv.Itoa(int(a.ID)),
		"remote_id":   strconv.Itoa(int(a.RemoteID)),
		"parent_id":   strconv.Itoa(int(a.ParentID)),
		"course_code": a.CourseCode,
		"title":       a.Title,
		"type_name":   a.TypeName,
		"deadline":    a.Deadline.Format(time.DateOnly),
		"todo":        a.Todo,
		"status_name": a.StatusName,
		"link":        a.Link,
		"priority":    a.Priority,
	}
}

func Get_Local_Assignment_byId(id uint, db *gorm.DB) (*LocalAssignment, error) {
	assignment := &LocalAssignment{}
	err := db.Preload("Course").
		Preload("Type").
		Preload("Status").
		Where("id = ?", id).
		First(assignment).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignment, nil
}
