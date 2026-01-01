package models

import (
	"strconv"

	"gorm.io/gorm"

	"unipilot/internal/errors"
)

// Note represents the note stored in the remote database
type Note struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index"`
	CourseID   uint   `gorm:"not null;index"`
	CourseCode string `gorm:"index"`
	Title      string `gorm:"not null"`
	Subject    string `gorm:"not null"`
	Content    string
	Videos     string
	ParentID   uint `gorm:"default:0"`

	User     User   `gorm:"foreignKey:UserID;references:ID"`
	Course   Course `gorm:"foreignKey:CourseID;references:ID"`
	Parent   *Note  `gorm:"foreignKey:ParentID;references:ID"`
	Children []Note `gorm:"foreignKey:ParentID"`
}

// LocalNote represents a note in the local database
type LocalNote struct {
	Note
	RemoteID       uint `gorm:"unique"`
	RemoteCourseID uint

	UserID uint        `gorm:"-"`
	Course LocalCourse `gorm:"foreignKey:CourseID;references:ID"`
}

func (n *Note) ToMap() map[string]string {
	return map[string]string{
		"id":          strconv.Itoa(int(n.ID)),
		"user_id":     strconv.Itoa(int(n.UserID)),
		"course_id":   strconv.Itoa(int(n.CourseID)),
		"course_code": n.CourseCode,
		"title":       n.Title,
		"subject":     n.Subject,
		"content":     n.Content,
		"videos":      n.Videos,
	}
}

func (n *LocalNote) ToMap() map[string]string {
	nMap := n.Note.ToMap()
	nMap["remote_id"] = strconv.Itoa(int(n.RemoteID))
	nMap["remote_course_id"] = strconv.Itoa(int(n.RemoteCourseID))
	return nMap
}

func (n *Note) ToLocal() *LocalNote {
	innerN := &Note{
		Title:      n.Title,
		Subject:    n.Subject,
		Content:    n.Content,
		Videos:     n.Videos,
		CourseCode: n.CourseCode,
	}
	return &LocalNote{
		Note:           *innerN,
		RemoteID:       n.ID,
		RemoteCourseID: n.CourseID,
	}
}

func (n *LocalNote) ToRemote() *Note {
	return &Note{
		UserID:     n.UserID,
		CourseID:   n.RemoteCourseID,
		CourseCode: n.CourseCode,
		Title:      n.Title,
		Subject:    n.Subject,
		Content:    n.Content,
		Videos:     n.Videos,
	}
}
func GetNote(id uint, db *gorm.DB) (*Note, error) {
	var note Note
	err := db.First(&note, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &note, nil
}
func GetLNote(id uint, db *gorm.DB) (*LocalNote, error) {
	var note LocalNote
	err := db.First(&note, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &note, nil
}

func GetNotes(userID uint, db *gorm.DB) ([]Note, error) {
	var notes []Note
	err := db.Where("user_id = ?", userID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetLNotes(userID uint, db *gorm.DB) ([]LocalNote, error) {
	var notes []LocalNote
	err := db.Where("user_id = ?", userID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetNotesByCourse(courseID uint, db *gorm.DB) ([]Note, error) {
	var notes []Note
	err := db.Where("course_id = ?", courseID).Order("created_at DESC").Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetLNotesByCourse(courseID uint, db *gorm.DB) ([]LocalNote, error) {
	var notes []LocalNote
	err := db.Where("course_id = ?", courseID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func DeleteNote(id uint, db *gorm.DB) error {
	err := db.Delete(&Note{}, "id = ?", id).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}
