package note

import (
	"strconv"

	"gorm.io/gorm"

	"unipilot/internal/errors"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
)

// RemoteNote represents the note stored in the remote database (Notion)
type Note struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	CourseCode string `json:"course_code"`
	Title      string `json:"title"`
	Subject    string `json:"subject"`
	Content    string `json:"content"`
	Videos     string `json:"videos"`

	User   user.User     `gorm:"foreignKey:UserID;references:ID"`
	Course course.Course `gorm:"foreignKey:CourseCode;references:Code"`
}

func (n *Note) ToMap() map[string]string {
	return map[string]string{
		"id":          strconv.Itoa(int(n.ID)),
		"user_id":     strconv.Itoa(int(n.UserID)),
		"title":       n.Title,
		"subject":     n.Subject,
		"content":     n.Content,
		"videos":      n.Videos,
		"course_code": n.CourseCode,
	}
}
func Get_Note_byID(id, user_id uint, db *gorm.DB) (*Note, error) {
	note := &Note{}
	err := db.Preload("User").
		Preload("Course", "user_id = ?", user_id).
		Where("id = ?", id).
		First(note).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return note, nil
}

func DeleteNote(id uint, db *gorm.DB) error {
	err := db.Delete(&Note{}, "id = ?", id).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}
