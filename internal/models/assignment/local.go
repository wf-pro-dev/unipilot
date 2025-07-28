package assignment

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending" // Needs to be synced
	SyncStatusSynced  SyncStatus = "synced"  // Already synced
)

type LocalAssignment struct {
	gorm.Model
	RemoteID   uint   `gorm:"unique"` // Empty until synced
	NotionID   string `gorm:"unique"` // Empty until synced
	Title      string `gorm:"not null"`
	Todo       string
	Deadline   time.Time  `gorm:"not null;index"`
	Link       string     `gorm:"default:https://acconline.austincc.edu/ultra/stream"`
	CourseCode string     `gorm:"not null;index"`
	TypeName   string     `gorm:"not null"`
	StatusName string     `gorm:"not null"`
	Priority   string     `gorm:"default:medium"`
	Completed  bool       `gorm:"default:false"`
	SyncStatus SyncStatus `gorm:"not null;default:'pending'"`

	Course course.LocalCourse           `gorm:"foreignKey:CourseCode;references:Code"`
	Type   models.LocalAssignmentType   `gorm:"foreignKey:TypeName;references:Name"`
	Status models.LocalAssignmentStatus `gorm:"foreignKey:StatusName;references:Name"`
}

func (a *LocalAssignment) ToMap() map[string]string {
	return map[string]string{
		"id":          strconv.Itoa(int(a.ID)),
		"remote_id":   strconv.Itoa(int(a.RemoteID)),
		"course_code": a.CourseCode,
		"title":       a.Title,
		"type_name":   a.TypeName,
		"deadline":    a.Deadline.Format(time.DateOnly),
		"todo":        a.Todo,
		"status_name": a.StatusName,
		"link":        a.Link,
		"notion_id":   a.NotionID,
		"priority":    a.Priority,
		"completed":   strconv.FormatBool(a.Completed),
		"sync_status": string(a.SyncStatus),
	}
}

func Get_Local_Assignment_byId(id uint, db *gorm.DB) (*LocalAssignment, error) {
	assignment := &LocalAssignment{}
	err := db.Preload("Course").
		Preload("Type").
		Preload("Status").
		Where("id = ?", id).
		First(assignment).Error

	if err != nil {
		return nil, err
	}
	return assignment, nil
}

func GetAssignmentsbyCourse(course_code string, columns []string, filters []Filter, up_to_date bool, db *gorm.DB) {

	col_length := 15
	query := fmt.Sprintf("SELECT %s FROM local_assignments WHERE course_code='%v' AND deleted_at is NULL", strings.Join(columns, ","), course_code)

	for _, filter := range filters {

		if filter.Column == "deadline" {
			deadline, err := time.Parse(time.DateOnly, filter.Value)
			if err != nil {
				log.Fatal(err)
			}

			t := time.Now()
			deadline_str := deadline.Format("2006-01-02 15:04:05") + t.Format("-07:00")
			query += fmt.Sprintf(" AND deadline = '%v'", deadline_str)
		} else {
			query += fmt.Sprintf(" AND %s='%v'", filter.Column, filter.Value)
		}
	}

	if up_to_date {
		today := time.Now().Format(time.DateOnly)
		query += fmt.Sprintf(" AND deadline >= '%v'", today)
	}
	query += " ORDER BY deadline ASC"
	assignments := []LocalAssignment{}
	err := db.Raw(query).Scan(&assignments).Error
	if err != nil {
		log.Fatal(err)
	}

	if len(assignments) == 0 {
		fmt.Println("No assignments found")
		os.Exit(0)
	}

	// Create column headers based on requested columns
	headers := make([]string, len(columns))
	for i, col := range columns {
		// Convert column names to display headers
		switch col {
		case "id":
			headers[i] = "ID"
		case "type_name":
			headers[i] = "Type"
		case "deadline":
			headers[i] = "Deadline"
		case "title":
			headers[i] = "Title"
		case "todo":
			headers[i] = "Todo"
		case "course_code":
			headers[i] = "Course Code"
		case "status_name":
			headers[i] = "Status"
		default:
			headers[i] = col
		}
	}

	// Print top border
	fmt.Print("┌")
	for range columns {
		fmt.Printf("%-*s┬", col_length, strings.Repeat("-", col_length+2))
	}
	fmt.Println("")

	// Print header row
	fmt.Print("│")
	for _, header := range headers {
		fmt.Printf(" %-*s │", col_length, header)
	}
	fmt.Println("")

	// Print separator
	fmt.Print("├")
	for range columns {
		fmt.Printf("%-*s┼", col_length, strings.Repeat("-", col_length+2))
	}
	fmt.Println("")

	// Print data rows
	for _, assignment := range assignments {
		obj_assign := assignment.ToMap()
		fmt.Print("│")
		for _, col := range columns {
			value := obj_assign[col]
			if col == "deadline" {
				value = value[:10]
			}

			// Truncate or pad to exactly 10 characters
			if len(value) > 15 && len(columns) > 2 {
				value = value[:12] + "..."
			}
			fmt.Printf(" %-*s │", col_length, value)
		}
		fmt.Println("")
	}

	// Print bottom border
	fmt.Print("└")
	for range columns {
		fmt.Printf("%-*s┴", col_length, strings.Repeat("-", col_length+2))
	}
	fmt.Println("")
}
