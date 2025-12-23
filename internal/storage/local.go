package storage

import (
	"log"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/aimessage"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/document"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"

	"gorm.io/gorm"
)

func InitializeSchema(db *gorm.DB) error {
	log.Println("Initializing schema")
	// Run migrations
	err := db.AutoMigrate(
		&course.LocalCourse{},
		&models.LocalAssignmentType{},
		&models.LocalAssignmentStatus{},
		&assignment.LocalAssignment{},
		&models.LocalUpdate{},
		&document.LocalDocument{},
		&note.LocalNote{},
		&notifications.LocalNotification{},
		&aimessage.LocalAiMessage{},
	)

	if err != nil {
		return errors.Wrap(err, errors.DBQueryFailed, "Failed to migrate schema")
	}

	types := []*models.LocalAssignmentType{
		{ID: 1, Name: "HW", Color: "yellow", NotionID: "Vn}Z"},
		{ID: 2, Name: "Exam", Color: "red", NotionID: "oiNS"},
	}

	// Assignment statuses
	statuses := []*models.LocalAssignmentStatus{
		{ID: 1, Name: "Not started", Color: "default", NotionID: "3aa77cf8-c39e-4c7b-b7d2-ab15ae43ff23"},
		{ID: 2, Name: "In progress", Color: "blue", NotionID: "97903420-1e83-4b3a-9eaf-a904354c968b"},
		{ID: 3, Name: "Done", Color: "green", NotionID: "2fef8044-d8d7-4fcf-a3ee-393a1d558e94"},
	}

	for _, t := range types {
		if err := db.Where("id = ?", t.ID).First(&models.LocalAssignmentType{}).Error; err != nil {
			err = db.Create(t).Error
			if err != nil {
				return errors.HandleDBCreateError(err)
			}
		}
	}

	for _, status := range statuses {
		if err := db.Where("id = ?", status.ID).First(&models.LocalAssignmentStatus{}).Error; err != nil {
			err = db.Create(status).Error
			if err != nil {
				return errors.HandleDBCreateError(err)
			}
		}
	}

	return nil
}
