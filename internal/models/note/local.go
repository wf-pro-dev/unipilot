package note

import (
	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models/course"
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

func GetLocalCourseNotes(courseCode string, db *gorm.DB) ([]LocalNote, error) {
	var notes []LocalNote
	err := db.Where("course_code = ?", courseCode).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}
