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
	Completed  bool   `gorm:"default:false"`

	User      user.User               `gorm:"foreignKey:UserID;references:ID"`
	Course    course.Course           `gorm:"foreignKey:CourseCode;references:Code"`
	Type      models.AssignmentType   `gorm:"foreignKey:TypeName;references:Name"`
	Status    models.AssignmentStatus `gorm:"foreignKey:StatusName;references:Name"`
	Documents []document.Document     `gorm:"foreignKey:AssignmentID;references:ID"`
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

func Get_Assignment_byId(id, user_id uint, db *gorm.DB) (*Assignment, error) {
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
		Where("local_id = ?", id).
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

// ToMap converts the Assignment struct to a map[string]string
// This maintains compatibility with the existing database operations
func (a *Assignment) ToMap() map[string]string {

	return map[string]string{
		"id":          strconv.Itoa(int(a.ID)),
		"local_id":    strconv.Itoa(int(a.LocalID)),
		"user_id":     strconv.Itoa(int(a.UserID)),
		"local_id":    strconv.Itoa(int(a.LocalID)),
		"notion_id":   a.NotionID,
		"type":        a.TypeName,
		"deadline":    a.Deadline.Format(time.DateOnly),
		"title":       a.Title,
		"todo":        a.Todo,
		"deadline":    a.Deadline.Format(time.RFC3339),
		"course_code": a.CourseCode,
		"type":        a.TypeName,
		"status":      a.StatusName,
		"link":        a.Link,
		"priority":    a.Priority,
		"completed":   strconv.FormatBool(a.Completed),
		"created_at":  a.CreatedAt.Format(time.RFC3339),
		"updated_at":  a.UpdatedAt.Format(time.RFC3339),
		"notion_id":   a.NotionID,
	}
}

func (a *Assignment) Add(db *gorm.DB) (err error) {

	assignment := a.ToMap()

	delete(assignment, "id")

	err = db.Create(a).Error

	if err != nil {
		log.Fatalln("Error adding assignment to database: ", err)
		return err
	}

	notion_id, err_notion := a.Add_Notion()

	if err_notion != nil {
		log.Fatalln("Error adding assignment to Notion: ", err_notion)
		return err_notion
	}

	var lastVal int
	err = db.Raw("SELECT MAX(id) FROM assignements").Scan(&lastVal).Error
	if err != nil {
		log.Fatal(err)
	}
	err = db.Model(&Assignment{}).Where("id = ?", lastVal).Update("notion_id", notion_id).Error

	if err != nil {
		log.Fatalln("Error updating assignment: ", err)
		return err
	}

	return nil
}

/*
func (a *Assignment) Update(col, value string, db *gorm.DB) (err error) {

		err = db.Model(&Assignment{}).Where("id = ?", a.ID).Update(col, value).Error

		if err != nil {
			log.Fatalln("Error updating assignment in database: ", err)
			return err
		}

		assignment := a.ToMap()
		assignment[col] = value

		if col == "course_code" {
			c , _ := course.Get_Course_byCode(value, db)
			value = c.NotionID
		}

		var obj map[string]string

		if col == "status" {
			obj = models.Get_AssignmentStatus_byName(value, db).ToMap()
		} else {
			obj = models.Get_AssignmentType_byName(value, db).ToMap()
		}

		err = a.Update_Notion(col, value, obj)

		if err != nil {
			log.Fatalln("Error updating assignment to Notion: ", err)
			return err
		}

		return nil
	}
*/
func (a *Assignment) Delete(db *gorm.DB) (err error) {

	err = db.Delete(a).Error

	if err != nil {
		log.Fatalln(err)
	}

	err = a.Delete_Notion()
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

// Document-related methods

// GetDocuments retrieves all documents for this assignment
func (a *Assignment) GetDocuments(db *gorm.DB) ([]document.Document, error) {
	return document.GetDocumentsByAssignment(a.ID, a.UserID, db)
}

// GetLatestDocuments retrieves only the latest versions of documents for this assignment
func (a *Assignment) GetLatestDocuments(db *gorm.DB) ([]document.Document, error) {
	return document.GetLatestVersions(a.ID, a.UserID, db)
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
