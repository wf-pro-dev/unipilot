package note

import (
	"strconv"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// RemoteNote represents the note stored in the remote database (Notion)
type Note struct {
	gorm.Model
	LocalID    uint   `json:"local_id"`
	UserID     uint   `json:"user_id"`
	CourseCode string `json:"course_code"`
	Title      string `json:"title"`
	Subject    string `json:"subject"`
	Content    string `json:"content"`
	Keywords   string `json:"keywords"`
	Videos     string `json:"videos"`

	User   user.User     `gorm:"foreignKey:UserID;references:ID"`
	Course course.Course `gorm:"foreignKey:CourseCode;references:Code"`
}

func (n *Note) ToMap() map[string]string {
	return map[string]string{
		"local_id":    strconv.Itoa(int(n.LocalID)),
		"user_id":     strconv.Itoa(int(n.UserID)),
		"title":       n.Title,
		"subject":     n.Subject,
		"content":     n.Content,
		"keywords":    n.Keywords,
		"videos":      n.Videos,
		"course_code": n.CourseCode,
	}
}
