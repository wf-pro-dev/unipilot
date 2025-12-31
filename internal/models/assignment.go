package models

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/errors"

	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"
)

// Assignment represents a homework or exam assignment
type Assignment struct {
	gorm.Model
	UserID uint

	// Common fields
	Title      string `gorm:"not null"`
	Type       string `gorm:"not null"`
	Status     string `gorm:"not null"`
	Todo       string
	Deadline   time.Time `gorm:"not null"`
	Link       string    `gorm:"default:https://acconline.austincc.edu/ultra/stream"`
	CourseID   uint      `gorm:"not null;index"`
	CourseCode string    `gorm:"index"`
	Priority   string    `gorm:"default:medium"`
	ParentID   uint      `gorm:"default:0"`

	// Relationships
	User      User         `gorm:"foreignKey:UserID;references:ID"`
	Course    Course       `gorm:"foreignKey:CourseID;references:ID"`
	Documents []Document   `gorm:"foreignKey:AssignmentID;references:ID"`
	Parent    *Assignment  `gorm:"foreignKey:ParentID;references:ID"`
	Children  []Assignment `gorm:"foreignKey:ParentID"`
}

type LocalAssignment struct {
	Assignment
	RemoteID       uint `gorm:"unique"`
	RemoteCourseID uint

	UserID uint `gorm:"-"`

	Course    LocalCourse       `gorm:"foreignKey:CourseID;references:ID"`
	Documents []LocalDocument   `gorm:"foreignKey:AssignmentID;references:ID"`
	Parent    *LocalAssignment  `gorm:"foreignKey:ParentID;references:ID"`
	Children  []LocalAssignment `gorm:"foreignKey:ParentID"`
}

// ToMap converts the Assignment struct to a map[string]string
func (a *Assignment) ToMap() map[string]string {

	return map[string]string{
		"id":               strconv.Itoa(int(a.ID)),
		"user_id":          strconv.Itoa(int(a.UserID)),
		"course_id":        strconv.Itoa(int(a.CourseID)),
		"parent_id":        strconv.Itoa(int(a.ParentID)),
		"type":             a.Type,
		"deadline":         a.Deadline.Format(time.DateOnly),
		"title":            a.Title,
		"todo":             a.Todo,
		"remote_course_id": strconv.Itoa(int(a.CourseID)),
		"course_code":      a.CourseCode,
		"status":           a.Status,
		"link":             a.Link,
		"priority":         a.Priority,
		"created_at":       a.CreatedAt.Format(time.RFC3339),
		"updated_at":       a.UpdatedAt.Format(time.RFC3339),
	}
}

func (a *Assignment) ToLocal() *LocalAssignment {
	assignment := &Assignment{
		Title:      a.Title,
		Todo:       a.Todo,
		Deadline:   a.Deadline,
		CourseCode: a.CourseCode,
		Type:       a.Type,
		Status:     a.Status,
		Priority:   a.Priority,
		Link:       a.Link,
		ParentID:   a.ParentID,
	}
	return &LocalAssignment{
		Assignment:     *assignment,
		RemoteID:       a.ID,
		RemoteCourseID: a.CourseID,
	}
}

func (a *LocalAssignment) ToRemote() *Assignment {
	return &Assignment{
		UserID:     a.UserID,
		Title:      a.Title,
		Todo:       a.Todo,
		Deadline:   a.Deadline,
		CourseID:   a.RemoteCourseID,
		CourseCode: a.CourseCode,
		Type:       a.Type,
		Status:     a.Status,
		Priority:   a.Priority,
		Link:       a.Link,
		ParentID:   a.ParentID,
	}
}

func (la *LocalAssignment) ToMap() map[string]string {
	laMap := la.Assignment.ToMap()
	laMap["remote_id"] = strconv.Itoa(int(la.RemoteID))
	return laMap
}

// NewAssignment creates a new Assignment by prompting user for input
// This is equivalent to the createAssign function but returns a struct
func NewAssignment() *Assignment {

	fmt.Println("===== Creating new Assignement =====")

	assignment := &Assignment{}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("The type (HW or Exam): ")
	scanner.Scan()
	assignment.Type = scanner.Text()

	fmt.Printf("The deadline (yyyy-mm-dd): ")
	scanner.Scan()
	deadline, err := time.Parse(time.DateOnly, scanner.Text())
	if err != nil {
		log.Fatal("Error parsing deadline: ", err)
	}
	assignment.Deadline = deadline

	fmt.Printf("The title: ")
	scanner.Scan()
	assignment.Title = scanner.Text()

	fmt.Printf("The todo: ")
	scanner.Scan()
	assignment.Todo = scanner.Text()

	// Get course code from current directory name
	pwd := os.Getenv("PWD")
	cmd := exec.Command("basename", pwd)
	output, _ := cmd.CombinedOutput()
	assignment.CourseCode = strings.TrimSpace(string(output))

	assignment.Link = "https://acconline.austincc.edu/ultra/stream"

	return assignment
}

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
