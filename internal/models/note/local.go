package note

import (
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

// LocalNote represents the note stored in the local database
type LocalNote struct {
	gorm.Model
	RemoteID   uint   `gorm:"unique" json:"remote_id"`
	CourseCode string `gorm:"not null" json:"course_code"`
	Title      string `gorm:"not null" json:"title"`
	Subject    string `gorm:"not null" json:"subject"`
	Content    string `gorm:"type:text" json:"content"`
	Videos     string `gorm:"type:text" json:"videos"`

	Course course.LocalCourse `gorm:"foreignKey:CourseCode;references:Code"`
}
