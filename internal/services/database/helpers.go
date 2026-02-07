package database

import (
	"context"
	"fmt"
	"log"

	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/services/fileops"
)

// GetUser retrieves a user by ID
func (h *Database) GetUser(id string) (*models.User, error) {
	var u models.User
	err := h.db.First(&u, id).Error
	return &u, err
}

// GetLAssignment retrieves an assignment by ID
func (h *Database) GetLAssignment(id string) (*models.LocalAssignment, error) {
	assignment, err := models.GetLAssignment(id, h.db)
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return assignment, nil
}

// GetAssignments retrieves all assignments for a user
func (h *Database) GetAssignments() ([]models.LocalAssignment, error) {
	var LocalAssignment []models.LocalAssignment
	err := h.db.Preload("Course").Order("deadline DESC").Order("created_at DESC").Find(&LocalAssignment).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalAssignment, nil
}

// DeleteAssignment deletes an assignment
func (h *Database) DeleteAssignment(assignment *models.LocalAssignment) error {
	err := h.db.Delete(assignment).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetCourses retrieves all courses for a user
func (h *Database) GetCourses() ([]models.LocalCourse, error) {
	var LocalCourse []models.LocalCourse
	err := h.db.Order("start_date DESC").Find(&LocalCourse).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalCourse, nil
}

// CreateCourse creates a new course
func (h *Database) CreateCourse(course *models.Course) error {

	err := h.db.Create(course).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// UpdateCourse updates an existing course
func (h *Database) UpdateCourse(LocalCourse *models.LocalCourse, column, value string) error {
	// Only update the assignment fields, not the related course data
	err := h.db.Exec(fmt.Sprintf("UPDATE local_courses SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalCourse.ID)).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// DeleteCourse deletes a course
func (h *Database) DeleteCourse(course *models.LocalCourse) error {
	err := h.db.Delete(course).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// GetNotes returns all notes for the current user
func (h *Database) GetNotes() ([]models.LocalNote, error) {
	var LocalNote []models.LocalNote
	err := h.db.Preload("Course").Find(&LocalNote).Order("created_at DESC").Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}
	return LocalNote, nil
}

// CreateNote creates a new note
func (h *Database) CreateNote(note *models.LocalNote) error {
	err := h.db.Create(note).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// UpdateNote updates an existing note
func (h *Database) UpdateNote(LocalNote *models.LocalNote, column, value string) error {
	err := h.db.Exec(fmt.Sprintf("UPDATE local_notes SET %s = '%s', updated_at = CURRENT_TIMESTAMP WHERE id = '%d'", column, value, LocalNote.ID)).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

// DeleteNote deletes a note
func (h *Database) DeleteNote(note *models.LocalNote) error {
	err := h.db.Delete(note).Error
	if err != nil {
		return errors.HandleDBWriteError(err)
	}
	return nil
}

func (h *Database) CreateDocument(ctx context.Context, document *models.LocalDocument) error {

	log.Println("Creating document", document.ID, document.FileSize)
	var err error

	db := h.db.Debug()

	if document.HasLocalFile {
		// Upload the document locally
		err = fileops.WriteDocument(document, db)
		if err != nil {
			return errors.Wrap(err, errors.FSWriteFailed, "Failed to write document on disk")
		}
	}

	if err = document.Validate(h.db); err != nil {
		return err
	}

	if err = db.Create(document).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// Saving a message from AI SDK
func (h *Database) SaveUIMessage(assignmentID string, vercelMessage map[string]interface{}) error {
	message, err := models.FromUIMessage(assignmentID, vercelMessage)
	if err != nil {
		return errors.Wrap(err, errors.ValidationInvalid, "Failed to create message from UI message")
	}
	err = h.db.Create(message).Error
	if err != nil {
		return errors.HandleDBCreateError(err)
	}
	return nil
}

// Retrieving conversation history
func (h *Database) GetConversationHistory(assignmentID string) ([]models.LocalAiMessage, error) {
	var dbMessages []models.LocalAiMessage
	err := h.db.Where("assignment_id = ?", assignmentID).
		Order("created_at ASC").
		Find(&dbMessages).Error
	if err != nil {
		return nil, errors.HandleDBReadError(err)
	}

	uiMessages := make([]map[string]interface{}, len(dbMessages))
	for i, msg := range dbMessages {
		uiMessages[i], err = msg.ToUIMessage()
		if err != nil {
			return nil, errors.Wrap(err, errors.ProcDataProcessingFailed, "Failed to convert message to UI message")
		}
	}

	return dbMessages, nil
}
