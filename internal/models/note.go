package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"unipilot/internal/errors"
)

type BaseNote struct {
	Title    string `gorm:"not null" validate:"required,min=3,max=100"`
	Subject  string `gorm:"not null" validate:"required,min=3,max=100"`
	Content  string `gorm:"not null" validate:"max=50000"`
	Videos   string
	ParentID *string `gorm:"index;default:null"`

	CourseID string `gorm:"not null;index" validate:"required"`
}

// Note represents the note stored in the remote database
type Note struct {
	Base
	BaseNote
	UserID string `gorm:"not null;index" validate:"required"`

	User     *User   `gorm:"foreignKey:UserID;references:ID" validate:"-"`
	Course   *Course `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Parent   *Note   `gorm:"foreignKey:ParentID;references:ID" validate:"-"`
	Children []Note  `gorm:"foreignKey:ParentID" validate:"-"`
}

// LocalNote represents a note in the local database
type LocalNote struct {
	Base
	BaseNote
	SyncedAt *time.Time `gorm:"default:null"`

	Course *LocalCourse `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
}

func (n *Note) ToLocal() *LocalNote {

	return &LocalNote{
		BaseNote: n.BaseNote,
	}
}

func (n *LocalNote) ToRemote(userID string) *Note {

	return &Note{
		BaseNote: n.BaseNote,
	}
}

// END : CONVERSION FUNCTIONS

// START : VALIDATION FUNCTIONS

func (n *Note) Validate() error {
	if err := validator.New().Struct(n); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Note validation failed")
	}
	return nil
}

func (n *LocalNote) Validate() error {

	if err := validator.New().Struct(n); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "LocalNote validation failed")
	}
	return nil
}

// END : VALIDATION FUNCTIONS

func GetNote(id string, db *gorm.DB) (*Note, error) {
	var note Note
	err := db.First(&note, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &note, nil
}

func GetLNote(id string, db *gorm.DB) (*LocalNote, error) {
	var note LocalNote
	err := db.First(&note, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return &note, nil
}

func GetNotes(userID string, db *gorm.DB) ([]Note, error) {
	var notes []Note
	err := db.Where("user_id = ?", userID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetLNotes(userID string, db *gorm.DB) ([]LocalNote, error) {
	var notes []LocalNote
	err := db.Where("user_id = ?", userID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetNotesByIDs(noteIDs []string, db *gorm.DB) ([]*Note, error) {
	var notes []*Note
	err := db.Where(noteIDs).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func GetNoteContent(noteID string, db *gorm.DB) (string, error) {
	var content string
	err := db.Select("content").First(&content, noteID).Error
	if err != nil {
		return "", errors.HandleDBReadError(err)
	}
	return content, nil
}

func (c *Course) GetCourseNotes(db *gorm.DB) ([]Note, error) {
	var notes []Note
	err := db.Model(&c).Association("Notes").Find(&notes)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func (c *Course) GetCourseNoteIDs(db *gorm.DB) ([]string, error) {
	var noteIDs []string
	// Pluck extracts a single column
	err := db.Model(&Note{}).
		Where("course_id = ?", c.ID).
		Pluck("id", &noteIDs).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return noteIDs, nil
}

func (lc *LocalCourse) GetNotesByCourse(db *gorm.DB) ([]LocalNote, error) {
	var notes []LocalNote
	err := db.Where("course_id = ?", lc.ID).Find(&notes).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return notes, nil
}

func DeleteNote(id string, db *gorm.DB) error {
	err := db.Delete(&Note{}, "id = ?", id).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

func (n *Note) ClusterRoot() string {
	if n.Course.ClusterID != nil {
		return *n.Course.ClusterID
	}
	return n.Course.ID
}

func (n *Note) IsCopy() bool { return n.ParentID != nil }
