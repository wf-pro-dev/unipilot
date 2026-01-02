package models

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/errors"

	"github.com/go-playground/validator/v10"
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"
)

type BaseAssignment struct {
	Title      string    `gorm:"not null" validate:"required,min=3,max=100"`
	Type       string    `gorm:"not null" validate:"required,oneof=HW Group_project Exam Quiz Lab"`
	Status     string    `gorm:"not null" validate:"required,oneof=Not_started In_progress Done"`
	Todo       string    `gorm:"not null" validate:"max=1000"`
	Deadline   time.Time `gorm:"not null" validate:"required"`
	Link       string    `gorm:"default:https://acconline.austincc.edu/ultra/stream" validate:"url"`
	CourseID   uint      `gorm:"not null;index" validate:"required"`
	CourseCode string    `gorm:"index" validate:"required"`
	Priority   string    `gorm:"default:low" validate:"required,oneof=low medium high"`
	ParentID   uint      `gorm:"default:null"`
}

// Assignment represents a homework or exam assignment
type Assignment struct {
	gorm.Model
	BaseAssignment
	UserID uint `gorm:"not null" validate:"required"`

	// Relationships
	User      *User        `gorm:"foreignKey:UserID;references:ID" validate:"-"`
	Course    *Course      `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Documents []Document   `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
	Parent    *Assignment  `gorm:"foreignKey:ParentID;references:ID" validate:"-"`
	Children  []Assignment `gorm:"foreignKey:ParentID" validate:"-"`
}

type LocalAssignment struct {
	gorm.Model
	BaseAssignment
	RemoteID       uint `gorm:"unique;default:null"`
	RemoteCourseID uint `gorm:"default:null"`

	Course    *LocalCourse    `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Documents []LocalDocument `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
}

func (a *BaseAssignment) ToMap() map[string]string {
	return map[string]string{
		"title":       a.Title,
		"type":        a.Type,
		"status":      a.Status,
		"todo":        a.Todo,
		"deadline":    a.Deadline.Format(time.DateOnly),
		"course_id":   strconv.Itoa(int(a.CourseID)),
		"course_code": a.CourseCode,
		"priority":    a.Priority,
		"link":        a.Link,
		"parent_id":   strconv.Itoa(int(a.ParentID)),
	}
}

// ToMap converts the Assignment struct to a map[string]string
func (a *Assignment) ToMap() map[string]string {
	aMap := a.BaseAssignment.ToMap()
	aMap["id"] = strconv.Itoa(int(a.ID))
	aMap["user_id"] = strconv.Itoa(int(a.UserID))
	aMap["remote_course_id"] = strconv.Itoa(int(a.CourseID))
	return aMap
}

func (a *Assignment) ToLocal() *LocalAssignment {
	return &LocalAssignment{
		BaseAssignment: a.BaseAssignment,
		RemoteID:       a.ID,
		RemoteCourseID: a.CourseID,
	}
}

func (a *LocalAssignment) ToRemote() *Assignment {
	assignment := &Assignment{
		BaseAssignment: a.BaseAssignment,
	}
	return assignment
}

func (la *LocalAssignment) ToMap() map[string]string {
	laMap := la.BaseAssignment.ToMap()
	laMap["remote_id"] = strconv.Itoa(int(la.RemoteID))
	return laMap
}

// END: Conversion Functions

// START: GORM Hooks

func (a *Assignment) BeforeDelete(tx *gorm.DB) error {
	// Retrieve qdrantClient from transaction context
	qdrantClient, ok := tx.Get("qdrantClient")
	if !ok {
		return nil
	}

	client, ok := qdrantClient.(*qdrant.Client)
	if !ok {
		return nil
	}

	documents, err := GetDocumentsByAssignment(a.ID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting documents by assignment")
	}

	for _, document := range documents {
		err := tx.Delete(&document).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting document")
		}
	}

	// Delete the Qdrant collection for the assignment
	collectionName := GetQdrantCollectionName(a.ID)
	err = client.DeleteCollection(context.Background(), collectionName)
	if err != nil {
		return errors.Wrap(err, errors.QdrantFailed, "Error deleting Qdrant collection")
	}

	return nil
}

func (la *LocalAssignment) AfterDelete(tx *gorm.DB) error {

	documents, err := GetDocumentsByAssignment(la.ID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting documents by assignment")
	}

	for _, document := range documents {
		err := tx.Delete(&document).Error
		if err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting document")
		}
	}

	return nil
}

// END: GORM Hooks

// START: Validation Functions

func (a *Assignment) Validate() error {

	a.Title = strings.TrimRight(a.Title, " ")
	a.Title = strings.TrimLeft(a.Title, " ")

	a.Todo = strings.TrimRight(a.Todo, " ")
	a.Todo = strings.TrimLeft(a.Todo, " ")

	a.Link = strings.TrimSpace(a.Link)

	a.Priority = strings.TrimRight(a.Priority, " ")
	a.Priority = strings.TrimLeft(a.Priority, " ")

	a.CourseCode = strings.TrimRight(a.CourseCode, " ")
	a.CourseCode = strings.TrimLeft(a.CourseCode, " ")
	a.CourseCode = strings.ToUpper(a.CourseCode)

	a.Type = strings.TrimRight(a.Type, " ")
	a.Type = strings.TrimLeft(a.Type, " ")

	a.Status = strings.TrimRight(a.Status, " ")
	a.Status = strings.TrimLeft(a.Status, " ")

	if err := isValidTitle(a.Title); err != nil {
		return err
	}
	if err := isValidTodo(a.Todo); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(a); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Assignment Validation failed")
	}

	return nil
}

func isValidTitle(title string) error {

	// Remove dangerous patterns but allow most characters
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)data:`),
		regexp.MustCompile(`(?i)vbscript:`),
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(title) {
			return errors.Wrap(fmt.Errorf("title contains unsafe pattern"), errors.ValidationInvalid, "Title contains unsafe content")
		}
	}

	return nil
}

func isValidTodo(todo string) error {
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)data:`),
		regexp.MustCompile(`(?i)vbscript:`),
	}
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(todo) {
			return errors.Wrap(fmt.Errorf("todo contains unsafe pattern"), errors.ValidationInvalid, "Todo contains unsafe content")
		}
	}
	return nil
}

// END: Validation Functions

// GET Operation

func GetAssignment(id uint, db *gorm.DB) (*Assignment, error) {
	assignment := &Assignment{}
	err := db.First(&assignment, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignment, nil
}

func GetLAssignment(id uint, db *gorm.DB) (*LocalAssignment, error) {
	assignment := &LocalAssignment{}
	err := db.First(&assignment, id).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignment, nil
}

func GetAssignments(userID uint, db *gorm.DB) ([]Assignment, error) {
	var assignments []Assignment
	err := db.Where("user_id = ?", userID).Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

func GetLAssignments(db *gorm.DB) ([]LocalAssignment, error) {
	var assignments []LocalAssignment
	err := db.Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

func GetAssignmentsByCourse(courseID uint, db *gorm.DB) ([]Assignment, error) {
	var assignments []Assignment
	err := db.Where("course_id = ?", courseID).Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

func GetLAssignmentsByCourse(courseCode string, db *gorm.DB) ([]LocalAssignment, error) {
	var assignments []LocalAssignment
	err := db.Where("course_code = ?", courseCode).Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

// DELETE Operations

// CHECK Operations

func (a *Assignment) IsRoot() bool       { return a.ParentID == 0 }
func (la *LocalAssignment) IsRoot() bool { return la.ParentID == 0 }

func GetQdrantCollectionName(assignmentID uint) string {
	return fmt.Sprintf("unipilot-qdrant-db-%d", assignmentID)
}

func GetAssignmentDocumentIDsRAG(assignmentID uint, qdrantClient *qdrant.Client) ([]uint, error) {
	ctx := context.Background()
	collectionName := GetQdrantCollectionName(assignmentID)

	exists, err := qdrantClient.CollectionExists(ctx, collectionName)
	if err != nil || !exists {
		return nil, errors.Wrap(err, errors.QdrantCollectionNotFound, "Assignment collection could not be found")
	}
	// Retrive All Qdrant Points for that assignment
	points, err := qdrantClient.Scroll(context.Background(), &qdrant.ScrollPoints{
		CollectionName: fmt.Sprintf("unipilot-qdrant-db-%d", assignmentID),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.QdrantFailed, "Error listing Qdrant points")
	}

	// Step 3: Make a set of document IDs that are in the Qdrant
	type Set[E comparable] map[E]struct{} // Generic set type
	uploadedDocumentIDs := Set[uint]{}
	for _, point := range points {
		uploadedDocumentIDs[uint(point.Payload["document_id"].GetIntegerValue())] = struct{}{} // Add document ID to set
	}

	// Step 4: Flatten the set of uploaded document IDs
	uploadedDocumentIDsList := make([]uint, 0, len(uploadedDocumentIDs))
	for id := range uploadedDocumentIDs {
		uploadedDocumentIDsList = append(uploadedDocumentIDsList, id)
	}
	return uploadedDocumentIDsList, nil
}
