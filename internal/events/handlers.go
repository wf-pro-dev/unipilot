package events

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"unipilot/internal/errors"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/notifications"
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

	db := h.DB

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ar AssignmentResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal assignment data")
		log.Printf("Error unmarshalling assignment: %v", wrappedErr)
		return
	}

	id, err := strconv.Atoi(ar.ID)
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcDataConversionFailed, "Failed to convert assignment ID to int")
		log.Printf("Error converting ID to int: %v", wrappedErr)
		return
	}

	deadline, err := time.Parse(time.RFC3339, ar.Deadline)
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcDataProcessingFailed, "Failed to parse assignment deadline")
		log.Printf("Error parsing deadline: %v", wrappedErr)
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
		wrappedErr := errors.HandleDBCreateError(err)
		log.Printf("Error creating assignment: %v", wrappedErr)
		return
	}

	tx.Commit()

}

func (h *Events) HandleAssignmentUpdate(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with update logic

	db := h.DB

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
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal assignment update data")
		log.Printf("Error parsing update: %v", wrappedErr)
		return
	}

	if err := tx.Model(&assignment.LocalAssignment{}).Where("remote_id = ?", update.ID).Update(update.Column, update.Value).Error; err != nil {
		wrappedErr := errors.HandleDBWriteError(err)
		log.Printf("Error updating assignment %s with %s = %s: %v", update.ID, update.Column, update.Value, wrappedErr)
		tx.Rollback()
		return
	}

	var a assignment.LocalAssignment
	err := tx.Model(&assignment.LocalAssignment{}).Where("remote_id = ?", update.ID).First(&a).Error
	if err != nil {
		wrappedErr := errors.HandleDBReadError(err)
		log.Printf("Error getting assignment: %v", wrappedErr)
		return
	}

	tx.Commit()

}

func (h *Events) HandleAssignmentDelete(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with delete logic

	db := h.DB

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var ar AssignmentResponse
	if err := json.Unmarshal(data, &ar); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal assignment data")
		log.Printf("Error unmarshalling assignment: %v", wrappedErr)
		return
	}

	var a assignment.LocalAssignment
	if err := tx.Model(&assignment.LocalAssignment{}).Where("notion_id = ?", ar.NotionID).First(&a).Error; err != nil {
		tx.Rollback()
		wrappedErr := errors.HandleDBReadError(err)
		log.Printf("Error getting assignment: %v", wrappedErr)
		return
	}

	if err := tx.Where("notion_id = ?", ar.NotionID).Delete(&assignment.LocalAssignment{}).Error; err != nil {
		tx.Rollback()
		wrappedErr := errors.HandleDBWriteError(err)
		log.Printf("Error deleting assignment: %v", wrappedErr)
		return
	}

	tx.Commit()

}

func (h *Events) HandleFollow(data json.RawMessage, message string) {

	db := h.DB

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var n notifications.LocalNotification
	if err := json.Unmarshal(data, &n); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal notification data")
		log.Printf("Error unmarshalling notification: %v", wrappedErr)
		return
	}

	if err := tx.Create(&n).Error; err != nil {
		tx.Rollback()
		wrappedErr := errors.HandleDBCreateError(err)
		log.Printf("Error creating notification: %v", wrappedErr)
		return
	}

	tx.Commit()

}
