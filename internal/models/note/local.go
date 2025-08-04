package note

import (
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

// LocalNote represents the note stored in the local database
type LocalNote struct {
	gorm.Model
	CourseCode string
	Title      string `gorm:"not null"`
	Subject    string `gorm:"not null"`
	Content    string `gorm:"type:text"`
	Keywords   string `gorm:"type:text"`
	Videos     string `gorm:"type:text"`

	Course course.Course `gorm:"foreignKey:CourseCode;references:Code"`
}
