package sync

import (
	"unipilot/internal/models"
)

// FindRemoteAssignment finds a remote assignment by ID
func FindRemoteAssignment(remoteAssignments []models.Assignment, remoteID uint) *models.Assignment {
	for _, remoteEntity := range remoteAssignments {
		if remoteEntity.ID == remoteID {
			return &remoteEntity
		}
	}
	return nil
}

// FindRemoteCourse finds a remote course by ID
func FindRemoteCourse(remoteCourses []models.Course, remoteID uint) *models.Course {
	for _, remoteEntity := range remoteCourses {
		if remoteEntity.ID == remoteID {
			return &remoteEntity
		}
	}
	return nil
}

// FindRemoteNote finds a remote note by ID
func FindRemoteNote(remoteNotes []models.Note, remoteID uint) *models.Note {
	for _, remoteEntity := range remoteNotes {
		if remoteEntity.ID == remoteID {
			return &remoteEntity
		}
	}
	return nil
}

// FindRemoteDocument finds a remote document by ID
func FindRemoteDocument(remoteDocuments []models.Document, remoteID uint) *models.Document {
	for _, remoteEntity := range remoteDocuments {
		if remoteEntity.ID == remoteID {
			return &remoteEntity
		}
	}
	return nil
}
