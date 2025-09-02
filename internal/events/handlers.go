package events

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/notifications"
	"unipilot/internal/storage"
)

type AssignmentResponse struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	Title      string `json:"title"`
	Todo       string `json:"todo"`
	Deadline   string `json:"deadline"`
	Link       string `json:"link"`
	CourseCode string `json:"course_code"`
	TypeName   string `json:"type"`
	StatusName string `json:"status"`
	NotionID   string `json:"notion_id"`
}

func (h *Events) HandleAssignmentCreate(data json.RawMessage, message string) {

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ar AssignmentResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		log.Printf("Error unmarshalling assignment: %v", err)
		return
	}

	id, err := strconv.Atoi(ar.ID)
	if err != nil {
		log.Printf("Error converting ID to int: %v", err)
		return
	}

	deadline, err := time.Parse(time.RFC3339, ar.Deadline)
	if err != nil {
		log.Printf("Error parsing deadline: %v", err)
		return
	}

	a := assignment.LocalAssignment{
		RemoteID:   uint(id),
		Title:      ar.Title,
		Todo:       ar.Todo,
		Deadline:   deadline,
		Link:       ar.Link,
		CourseCode: ar.CourseCode,
		TypeName:   ar.TypeName,
		StatusName: ar.StatusName,
	}
	if err := tx.Create(&a).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating assignment: %v", err)
		return
	}

	tx.Commit()

}

func (h *Events) HandleAssignmentUpdate(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with update logic

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var update struct {
		ID     string `json:"id"`
		Column string `json:"column"`
		Value  string `json:"value"`
	}

	if err := json.Unmarshal(data, &update); err != nil {
		log.Printf("Error parsing update: %v", err)
		return
	}

	if err := tx.Model(&assignment.LocalAssignment{}).Where("remote_id = ?", update.ID).Update(update.Column, update.Value).Error; err != nil {
		fmt.Printf("Error updating assignment %s with %s = %s\n", update.ID, update.Column, update.Value)
		tx.Rollback()
		panic(err)
	}

	var a assignment.LocalAssignment
	err = tx.Model(&assignment.LocalAssignment{}).Where("remote_id = ?", update.ID).First(&a).Error
	if err != nil {
		log.Printf("Error getting assignment: %v", err)
		return
	}

	tx.Commit()

}

func (h *Events) HandleAssignmentDelete(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with delete logic

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ar AssignmentResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		log.Printf("Error unmarshalling assignment: %v", err)
		return
	}

	var a assignment.LocalAssignment
	if err := tx.Model(&assignment.LocalAssignment{}).Where("notion_id = ?", ar.NotionID).First(&a).Error; err != nil {
		tx.Rollback()
		log.Printf("Error getting assignment: %v", err)
		return
	}

	if err := tx.Where("notion_id = ?", ar.NotionID).Delete(&assignment.LocalAssignment{}).Error; err != nil {
		tx.Rollback()
		log.Printf("Error deleting assignment: %v", err)
		return
	}

	tx.Commit()

}

func (h *Events) HandleFollow(data json.RawMessage, message string) {

	db, _, err := storage.GetLocalDB()
	if err != nil {
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var n notifications.LocalNotification
	if err := json.Unmarshal(data, &n); err != nil {
		log.Printf("Error unmarshalling notification: %v", err)
		return
	}

	if err := tx.Create(&n).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating notification: %v", err)
		return
	}

	tx.Commit()

}
