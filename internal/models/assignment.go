package models

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"unipilot/internal/errors"

	"github.com/go-playground/validator/v10"
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BaseAssignment struct {
	Title    string          `gorm:"not null" validate:"required,min=3,max=100"`
	Type     string          `gorm:"not null" validate:"required,oneof=HW 'Group project' Exam Quiz Lab"`
	Status   string          `gorm:"not null" validate:"required,oneof='Not started' 'In progress' 'Done'"`
	Todo     string          `gorm:"not null" validate:"max=1000"`
	Deadline time.Time       `gorm:"not null" validate:"required"`
	Link     string          `gorm:"default:https://acconline.austincc.edu/ultra/stream" validate:"url"`
	CourseID datatypes.UUID  `gorm:"not null;index" validate:"required"`
	Priority string          `gorm:"default:low" validate:"required,oneof=low medium high"`
	ParentID *datatypes.UUID `gorm:"index"`
}

// Assignment represents a homework or exam assignment
type Assignment struct {
	Base
	BaseAssignment
	UserID datatypes.UUID `gorm:"not null" validate:"required"`

	// Relationships
	User      *User        `gorm:"foreignKey:UserID;references:ID" validate:"-"`
	Course    *Course      `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Documents []Document   `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
	Parent    *Assignment  `gorm:"foreignKey:ParentID;references:ID" validate:"-"`
	Children  []Assignment `gorm:"foreignKey:ParentID" validate:"-"`
}

type LocalAssignment struct {
	Base
	BaseAssignment
	SyncedAt *time.Time `gorm:"default:null"`

	Course    *LocalCourse    `gorm:"foreignKey:CourseID;references:ID" validate:"-"`
	Documents []LocalDocument `gorm:"foreignKey:AssignmentID;references:ID" validate:"-"`
}

func (a *Assignment) ToLocal() *LocalAssignment {
	return &LocalAssignment{
		BaseAssignment: a.BaseAssignment,
	}
}

func (a *LocalAssignment) ToRemote(userID datatypes.UUID) *Assignment {

	baseAssignment := a.BaseAssignment

	assignment := &Assignment{
		BaseAssignment: baseAssignment,
		UserID:         userID,
	}
	return assignment
}

// END: Conversion Functions

// START: GORM Hooks

func (a *Assignment) BeforeDelete(tx *gorm.DB) error {
	// Retrieve qdrantClient from transaction context
	client, ok := tx.Get("qdrantClient")
	if !ok {
		return nil
	}

	qdrantClient, ok := client.(*qdrant.Client)
	if !ok {
		return nil
	}

	documents, err := a.GetDocumentsByAssignment(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting documents by assignment")
	}

	if len(documents) > 0 {
		// Batch delete documents
		tx.Delete(documents)
	}

	// Delete the Qdrant collection for the assignment

	collectionName := GetQdrantCollectionName(a.ID)
	exists, err := qdrantClient.CollectionExists(context.Background(), collectionName)
	if err != nil {
		return errors.Wrap(err, errors.QdrantFailed, "Error checking if Qdrant collection exists")
	}
	if exists {
		err = qdrantClient.DeleteCollection(context.Background(), collectionName)
		if err != nil {
			return errors.Wrap(err, errors.QdrantFailed, "Error deleting Qdrant collection")
		}
	}

	return nil
}

func (la *LocalAssignment) BeforeDelete(tx *gorm.DB) error {

	documents, err := la.GetDocumentsByAssignment(tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting documents by assignment")
	}

	// Batch delete documents
	tx.Delete(documents)

	return nil
}

// END: GORM Hooks

// START: Validation Functions

func (ba *BaseAssignment) Validate() error {

	ba.Title = strings.TrimRight(ba.Title, " ")
	ba.Title = strings.TrimLeft(ba.Title, " ")

	ba.Todo = strings.TrimRight(ba.Todo, " ")
	ba.Todo = strings.TrimLeft(ba.Todo, " ")

	ba.Link = strings.TrimSpace(ba.Link)

	ba.Priority = strings.TrimRight(ba.Priority, " ")
	ba.Priority = strings.TrimLeft(ba.Priority, " ")

	ba.Type = strings.TrimRight(ba.Type, " ")
	ba.Type = strings.TrimLeft(ba.Type, " ")

	ba.Status = strings.TrimRight(ba.Status, " ")
	ba.Status = strings.TrimLeft(ba.Status, " ")

	if err := isValidTitle(ba.Title); err != nil {
		return err
	}
	if err := isValidTodo(ba.Todo); err != nil {
		return err
	}

	return nil
}

func (a *Assignment) Validate() error {

	if err := a.BaseAssignment.Validate(); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(a); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Assignment Validation failed")
	}

	return nil
}

func (la *LocalAssignment) Validate() error {
	if err := la.BaseAssignment.Validate(); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(la); err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "LocalAssignment Validation failed")
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

func GetAssignment(id datatypes.UUID, db *gorm.DB) (*Assignment, error) {
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

func GetAssignments(userID datatypes.UUID, db *gorm.DB) ([]Assignment, error) {
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

func GetAssignmentsByIDs(assignmentIDs []datatypes.UUID, db *gorm.DB) ([]*Assignment, error) {
	var assignments []*Assignment
	err := db.Where(assignmentIDs).Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

func (c *Course) GetCourseAssignments(db *gorm.DB) ([]*Assignment, error) {
	var assignments []*Assignment
	err := db.Model(&c).Association("Assignments").Find(&assignments)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

func (c *Course) GetCourseAssignmentIDs(db *gorm.DB) ([]uint, error) {
	var assignmentIDs []uint
	// Pluck extracts a single column
	err := db.Model(&Assignment{}).
		Where("course_id = ?", c.ID).
		Pluck("id", &assignmentIDs).Error

	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignmentIDs, nil
}

func (lc *LocalCourse) GetAssignmentsByCourse(db *gorm.DB) ([]LocalAssignment, error) {
	var assignments []LocalAssignment
	err := db.Where("course_code = ?", lc.Code).Find(&assignments).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignments, nil
}

// DELETE Operations

// CHECK Operations

func (a *Assignment) ClusterRoot() datatypes.UUID {
	if a.Course.ClusterID != nil {
		return *a.Course.ClusterID
	}
	return a.Course.ID
}

func (a *Assignment) IsCopy() bool       { return a.ParentID != nil }
func (la *LocalAssignment) IsRoot() bool { return la.ParentID == nil }

func GetQdrantCollectionName(assignmentID datatypes.UUID) string {
	return fmt.Sprintf("unipilot-qdrant-db-%d", assignmentID)
}

func GetAssignmentDocumentIDsRAG(assignmentID datatypes.UUID, qdrantClient *qdrant.Client) ([]uint, error) {
	ctx := context.Background()
	collectionName := GetQdrantCollectionName(assignmentID)

	exists, err := qdrantClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, errors.Wrap(err, errors.QdrantCollectionNotFound, "Assignment collection could not be found")
	}
	if !exists {
		return []uint{}, nil
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
