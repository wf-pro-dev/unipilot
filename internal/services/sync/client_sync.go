package sync

import (
	"log"
	"strconv"
	"time"

	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
)

// syncAssignment handles syncing an assignment from local to remote
func (m *Manager) syncAssignment(syncLog models.LocalSync, remoteAssignments []models.Assignment) error {
	switch syncLog.Action {
	case "create":
		return m.syncAssignmentCreate(syncLog)
	case "update", "delete":
		return m.syncAssignmentUpdate(syncLog, remoteAssignments)
	default:
		return errors.NewAppError(errors.SyncFailed, "Unknown action: "+syncLog.Action, nil)
	}
}

// syncAssignmentCreate creates a new assignment on the remote server
func (m *Manager) syncAssignmentCreate(syncLog models.LocalSync) error {
	var localAssignment models.LocalAssignment
	if err := m.db.Where("id = ?", syncLog.EntityID).First(&localAssignment).Error; err != nil {
		return errors.HandleDBReadError(err)
	}

	remoteAssignment := localAssignment.ToRemote()

	remoteID, err := client.CreateAssignment(remoteAssignment)
	if err != nil {
		return err
	}

	localAssignment.RemoteID = remoteID
	if err := m.db.Save(&localAssignment).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}

	return nil
}

// syncAssignmentUpdate updates or deletes an assignment on the remote server
func (m *Manager) syncAssignmentUpdate(syncLog models.LocalSync, remoteAssignments []models.Assignment) error {
	var localAssignment models.LocalAssignment
	if err := m.db.Unscoped().Where("id = ?", syncLog.EntityID).First(&localAssignment).Error; err != nil {
		return errors.HandleDBReadError(err)
	}

	remoteAssignment := FindRemoteAssignment(remoteAssignments, localAssignment.RemoteID)
	if remoteAssignment == nil {
		return errors.NewAppError(errors.SyncRemoteNotFound, "Remote assignment not found", nil)
	}

	// Check if the remote assignment has been updated since local change
	if remoteAssignment.UpdatedAt.Before(syncLog.UpdatedAt) {
		if localAssignment.RemoteID == 0 {
			return errors.NewAppError(errors.SyncInvalidRemoteID, "Remote assignment ID is 0", nil)
		}

		remoteID := strconv.Itoa(int(localAssignment.RemoteID))
		if err := client.UpdateAssignment(remoteID, syncLog.Column, syncLog.Value); err != nil {
			return err
		}
	} else {
		log.Printf("[Sync] Remote assignment has been updated: remote=%s, local=%s",
			remoteAssignment.UpdatedAt.Format(time.RFC3339),
			syncLog.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

// syncCourse handles syncing a course from local to remote
func (m *Manager) syncCourse(syncLog models.LocalSync, remoteCourses []models.Course) error {
	switch syncLog.Action {
	case "create":
		return m.syncCourseCreate(syncLog)
	case "update", "delete":
		return m.syncCourseUpdate(syncLog, remoteCourses)
	default:
		return errors.NewAppError(errors.SyncFailed, "Unknown action: "+syncLog.Action, nil)
	}
}

// syncCourseCreate creates a new course on the remote server
func (m *Manager) syncCourseCreate(syncLog models.LocalSync) error {
	var localCourse models.LocalCourse
	if err := m.db.Where("id = ?", syncLog.EntityID).First(&localCourse).Error; err != nil {
		return errors.HandleDBReadError(err)
	}

	remoteCourse := localCourse.ToRemote()

	remoteID, err := client.CreateCourse(remoteCourse)
	if err != nil {
		return err
	}

	localCourse.RemoteID = remoteID
	if err := m.db.Save(&localCourse).Error; err != nil {
		return errors.HandleDBWriteError(err)
	}

	return nil
}

// syncCourseUpdate updates or deletes a course on the remote server
func (m *Manager) syncCourseUpdate(syncLog models.LocalSync, remoteCourses []models.Course) error {
	var localCourse models.LocalCourse
	if err := m.db.Unscoped().Where("id = ?", syncLog.EntityID).First(&localCourse).Error; err != nil {
		return errors.HandleDBReadError(err)
	}

	remoteCourse := FindRemoteCourse(remoteCourses, localCourse.RemoteID)
	if remoteCourse == nil {
		return errors.NewAppError(errors.SyncRemoteNotFound, "Remote course not found", nil)
	}

	// Check if the remote course has been updated since local change
	if remoteCourse.UpdatedAt.Before(syncLog.UpdatedAt) {
		if localCourse.RemoteID == 0 {
			return errors.NewAppError(errors.SyncInvalidRemoteID, "Remote course ID is 0", nil)
		}

		remoteID := strconv.Itoa(int(localCourse.RemoteID))
		if err := client.UpdateCourse(remoteID, syncLog.Column, syncLog.Value); err != nil {
			return err
		}
	} else {
		log.Printf("[Sync] Remote course has been updated: remote=%s, local=%s",
			remoteCourse.UpdatedAt.Format(time.RFC3339),
			syncLog.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

// syncNote handles syncing a note from local to remote
func (m *Manager) syncNote(syncLog models.LocalSync, remoteNotes []models.Note) error {
	remoteNote := FindRemoteNote(remoteNotes, syncLog.EntityID)
	if remoteNote == nil {
		return errors.NewAppError(errors.SyncRemoteNotFound, "Remote note not found", nil)
	}

	log.Printf("[Sync] Syncing note: remote=%s, local=%s",
		remoteNote.UpdatedAt.Format(time.RFC3339),
		syncLog.UpdatedAt.Format(time.RFC3339))

	// Check if the remote note has been updated since local change
	if remoteNote.UpdatedAt.Before(syncLog.UpdatedAt) {
		remoteID := strconv.Itoa(int(syncLog.EntityID))
		if err := client.UpdateNote(remoteID, syncLog.Column, syncLog.Value); err != nil {
			return err
		}
	} else {
		log.Printf("[Sync] Remote note has been updated: remote=%s, local=%s",
			remoteNote.UpdatedAt.Format(time.RFC3339),
			syncLog.UpdatedAt.Format(time.RFC3339))
	}

	return nil
}

// syncDocument handles syncing a document from local to remote
func (m *Manager) syncDocument(syncLog models.LocalSync, remoteDocuments []models.Document) error {
	// Document syncing is handled differently (file uploads)
	// This is a placeholder for future implementation
	log.Printf("[Sync] Document syncing not yet fully implemented for sync log %d", syncLog.ID)
	return nil
}

// syncUser handles syncing user data from local to remote
func (m *Manager) syncUser(syncLog models.LocalSync) error {
	return client.UpdateUser(syncLog.Column, syncLog.Value)
}

// Helper functions to fetch remote data (delegated to client package)
func fetchRemoteAssignments() ([]models.Assignment, error) {
	return client.GetAssignments()
}

func fetchRemoteCourses() ([]models.Course, error) {
	return client.GetCourses()
}

func fetchRemoteNotes() ([]models.Note, error) {
	return client.GetNotes()
}

func fetchRemoteDocuments() ([]models.Document, error) {
	return client.GetDocuments()
}
