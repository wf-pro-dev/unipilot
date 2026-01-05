package sync

import (
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Migrator handles server-to-client synchronization (downloading remote data to local)
type Migrator struct {
	db *gorm.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// MigrateCourses downloads all courses from remote server and stores them locally
func (m *Migrator) MigrateCourses() error {
	remoteCourses, err := client.GetCourses()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote courses for migration")
	}

	if len(remoteCourses) == 0 {
		return nil
	}

	var localCourses []models.LocalCourse
	for _, rc := range remoteCourses {
		localCourse := rc.ToLocal()
		localCourses = append(localCourses, *localCourse)
	}

	if err := m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "remote_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "name", "code", "color", "location", "start_date", "end_date", "schedule", "credits", "semester", "instructor", "instructor_email", "parent_id"}),
	}).Create(&localCourses).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// MigrateAssignments downloads all assignments from remote server and stores them locally
func (m *Migrator) MigrateAssignments() error {
	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote assignments for migration")
	}
	if len(remoteAssignments) == 0 {
		return nil
	}

	var localAssignments []models.LocalAssignment
	for _, ra := range remoteAssignments {
		localAssignment := ra.ToLocal()
		localAssignments = append(localAssignments, *localAssignment)
	}

	if err := m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "remote_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "todo", "type", "status", "deadline", "link", "priority", "parent_id"}),
	}).Create(&localAssignments).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// MigrateDocuments downloads all documents from remote server and stores them locally
func (m *Migrator) MigrateDocuments() error {
	remoteDocuments, err := client.GetDocuments()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote documents for migration")
	}

	if len(remoteDocuments) == 0 {
		return nil
	}

	var localDocuments []models.LocalDocument
	for _, rd := range remoteDocuments {
		localDocument := rd.ToLocal()
		localDocuments = append(localDocuments, *localDocument)
	}

	if err := m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "remote_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "storage_key", "version", "parent_id", "parent_doc_id", "is_original", "has_local_file"}),
	}).Create(&localDocuments).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// MigrateNotes downloads all notes from remote server and stores them locally
func (m *Migrator) MigrateNotes() error {
	remoteNotes, err := client.GetNotes()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote notes for migration")
	}

	if len(remoteNotes) == 0 {
		return nil
	}

	var localNotes []models.LocalNote
	for _, rn := range remoteNotes {
		localNote := rn.ToLocal()
		localNotes = append(localNotes, *localNote)
	}

	if err := m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "remote_id"}},
		UpdateAll: true,
	}).Create(&localNotes).Error; err != nil {
		return errors.HandleDBCreateError(err)
	}

	return nil
}

// MigrateAll performs a full migration of all entities from server to client
func (m *Migrator) MigrateAll() error {
	if err := m.MigrateCourses(); err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to migrate courses")
	}

	if err := m.MigrateAssignments(); err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to migrate assignments")
	}

	if err := m.MigrateNotes(); err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to migrate notes")
	}

	if err := m.MigrateDocuments(); err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to migrate documents")
	}

	return nil
}
