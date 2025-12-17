package assignment

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/user"

	"gorm.io/gorm"
)

// Assignment represents a homework or exam assignment
type Assignment struct {
	gorm.Model
	UserID     uint
	LocalID    uint `gorm:"unique"`
	NotionID   string
	Title      string `gorm:"not null"`
	Todo       string
	Deadline   time.Time `gorm:"not null"`
	Link       string    `gorm:"default:https://acconline.austincc.edu/ultra/stream"`
	CourseCode string
	TypeName   string `gorm:"not null"`
	StatusName string `gorm:"not null"`
	Priority   string `gorm:"default:medium"`
	ParentID   uint   `gorm:"default:0"`

	User      user.User               `gorm:"foreignKey:UserID;references:ID"`
	Course    course.Course           `gorm:"foreignKey:CourseCode;references:Code"`
	Type      models.AssignmentType   `gorm:"foreignKey:TypeName;references:Name"`
	Status    models.AssignmentStatus `gorm:"foreignKey:StatusName;references:Name"`
	Documents []document.Document     `gorm:"foreignKey:AssignmentID;references:ID"`
	Parent    *Assignment             `gorm:"foreignKey:ParentID;references:ID"`
	Children  []*Assignment           `gorm:"foreignKey:ParentID;references:ID"`
}

type Filter struct {
	Column string
	Value  string
}

// NewAssignment creates a new Assignment by prompting user for input
// This is equivalent to the createAssign function but returns a struct
func NewAssignment() *Assignment {

	fmt.Println("===== Creating new Assignement =====")

	assignment := &Assignment{}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("The type (HW or Exam): ")
	scanner.Scan()
	assignment.TypeName = scanner.Text()

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

func Get_Assignment_byID(id, user_id uint, db *gorm.DB) (*Assignment, error) {
	assignment := &Assignment{}
	err := db.Preload("User").
		Preload("Course", "user_id = ?", user_id).
		Preload("Type").
		Preload("Status").
		Where("id = ?", id).
		First(assignment).Error

	if err != nil {
		return nil, err
	}
	return assignment, nil
}

func Get_Assignment_byLocalID(id, user_id uint, db *gorm.DB) (*Assignment, error) {
	assignment := &Assignment{}
	err := db.Preload("User").
		Preload("Course", "user_id = ?", user_id).
		Preload("Type").
		Preload("Status").
		Where("local_id = ? AND user_id = ?", id, user_id).
		First(assignment).Error

	if err != nil {
		return nil, err
	}
	return assignment, nil
}

func Get_Assignment_byNotionID(notion_id string, db *gorm.DB) (*Assignment, error) {

	assignment := &Assignment{}
	err := db.Where("notion_id = ?", notion_id).First(assignment).Error

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return assignment, nil
}

func (a *Assignment) GetCourseAssignment(db *gorm.DB) (*course.Course, error) {
	var c *course.Course
	err := db.Where("user_id = ? AND code = ?", a.UserID, a.CourseCode).First(&c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetAssignmentsbyCourse(courseCode string, db *gorm.DB) ([]Assignment, error) {
	var assignments []Assignment
	err := db.Where("course_code = ?", courseCode).Find(&assignments).Error
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

// ToMap converts the Assignment struct to a map[string]string
// This maintains compatibility with the existing database operations
func (a *Assignment) ToMap() map[string]string {

	return map[string]string{
		"id":          strconv.Itoa(int(a.ID)),
		"user_id":     strconv.Itoa(int(a.UserID)),
		"local_id":    strconv.Itoa(int(a.LocalID)),
		"parent_id":   strconv.Itoa(int(a.ParentID)),
		"notion_id":   a.NotionID,
		"type":        a.TypeName,
		"deadline":    a.Deadline.Format(time.DateOnly),
		"title":       a.Title,
		"todo":        a.Todo,
		"course_code": a.CourseCode,
		"status":      a.StatusName,
		"link":        a.Link,
		"priority":    a.Priority,
		"created_at":  a.CreatedAt.Format(time.RFC3339),
		"updated_at":  a.UpdatedAt.Format(time.RFC3339),
	}
}

func GetDocuments(assignmentID, userID uint, db *gorm.DB) ([]document.Document, error) {
	return document.GetDocumentsByAssignment(assignmentID, userID, db)
}

// GetLatestDocuments retrieves only the latest versions of documents for this assignment
func GetLatestDocuments(assignmentID, userID uint, db *gorm.DB) ([]document.Document, error) {
	return document.GetLatestVersions(assignmentID, userID, db)
}

// GetSupportDocuments retrieves only support documents for this assignment
func (a *Assignment) GetSupportDocuments(db *gorm.DB) ([]document.Document, error) {
	var documents []document.Document
	err := db.Preload("User").
		Where("assignment_id = ? AND user_id = ? AND type = ?", a.ID, a.UserID, document.DocumentTypeSupport).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get support documents: %w", err)
	}

	return documents, nil
}

// GetSubmissionDocuments retrieves only submission documents for this assignment
func (a *Assignment) GetSubmissionDocuments(db *gorm.DB) ([]document.Document, error) {
	var documents []document.Document
	err := db.Preload("User").
		Where("assignment_id = ? AND user_id = ? AND type = ?", a.ID, a.UserID, document.DocumentTypeSubmission).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get submission documents: %w", err)
	}

	return documents, nil
}

// GetDocumentStorageUsage returns total bytes used by documents in this assignment
func (a *Assignment) GetDocumentStorageUsage(db *gorm.DB) (int64, error) {
	var totalSize int64
	err := db.Model(&document.Document{}).
		Where("assignment_id = ? AND user_id = ?", a.ID, a.UserID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&totalSize).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate document storage: %w", err)
	}

	return totalSize, nil
}

func (a *Assignment) GetChildren(db *gorm.DB) ([]Assignment, error) {
	var children []Assignment
	err := db.Where("parent_id = ?", a.ID).
		Find(&children).Error
	if err != nil {
		return nil, err
	}
	return children, nil
}

func DeleteAssignment(assignment Assignment, tx *gorm.DB) error {

	documents, err := GetDocuments(assignment.ID, assignment.UserID, tx)
	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Error getting assignment documents from database")
	}

	// Step 4: Delete all documents related to the assignment
	for _, doc := range documents {
		if err := document.DeleteDocument(doc, tx); err != nil {
			return errors.Wrap(err, errors.DBQueryFailed, "Error deleting assignment document from database")
		}
	}

	// Step 5: Delete the assignment from the database
	if err := tx.Delete(&assignment).Error; err != nil {

		return errors.Wrap(err, errors.DBQueryFailed, "Error deleting assignment from database")
	}
	return nil
}
