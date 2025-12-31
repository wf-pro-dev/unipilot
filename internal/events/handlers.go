package events

import (
	"encoding/json"
	"log"

	"unipilot/internal/errors"
	"unipilot/internal/models"
)

func (h *Events) HandleAssignmentCreate(data json.RawMessage, message string) {

	db := h.DB

	var assign models.LocalAssignment
	if err := json.Unmarshal(data, &assign); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal assignment data")
		log.Printf("Error unmarshalling assignment: %v", wrappedErr)
		return
	}

	if err := db.Create(&assign).Error; err != nil {
		wrappedErr := errors.HandleDBCreateError(err)
		log.Printf("Error creating assignment: %v", wrappedErr)
		return
	}

}

func (h *Events) HandleAssignmentUpdate(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with update logic

	db := h.DB

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

	if err := db.Model(&models.LocalAssignment{}).Where("remote_id = ?", update.ID).Update(update.Column, update.Value).Error; err != nil {
		wrappedErr := errors.HandleDBWriteError(err)
		log.Printf("Error updating assignment %s with %s = %s: %v", update.ID, update.Column, update.Value, wrappedErr)
		return
	}

	var a models.LocalAssignment
	err := db.Model(&models.LocalAssignment{}).Where("remote_id = ?", update.ID).First(&a).Error
	if err != nil {
		wrappedErr := errors.HandleDBReadError(err)
		log.Printf("Error getting assignment: %v", wrappedErr)
		return
	}

}

func (h *Events) HandleAssignmentDelete(data json.RawMessage, message string) {
	// Similar to handleAssignmentCreate but with delete logic

	db := h.DB

	var assign models.LocalAssignment
	if err := json.Unmarshal(data, &assign); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal assignment data")
		log.Printf("Error unmarshalling assignment: %v", wrappedErr)
		return
	}

	if err := db.Delete(&assign).Error; err != nil {
		wrappedErr := errors.HandleDBWriteError(err)
		log.Printf("Error deleting assignment: %v", wrappedErr)
		return
	}

}

func (h *Events) HandleFollow(data json.RawMessage, message string) {

	db := h.DB

	var n models.LocalNotification
	if err := json.Unmarshal(data, &n); err != nil {
		wrappedErr := errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal notification data")
		log.Printf("Error unmarshalling notification: %v", wrappedErr)
		return
	}

	if err := db.Create(&n).Error; err != nil {
		wrappedErr := errors.HandleDBCreateError(err)
		log.Printf("Error creating notification: %v", wrappedErr)
		return
	}

}
