package sync

import (
	"log"
	"strconv"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

func Sync(db *gorm.DB) error {

	// Get remote assignments
	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.SyncFailed, "Failed to get remote assignments")
		log.Println("[Sync] Error getting remote assignments", wrappedErr)
		return wrappedErr
	}

	// Get remote courses
	remoteCourses, err := client.GetCourses()
	if err != nil {
		wrappedErr := errors.Wrap(err, errors.SyncFailed, "Failed to get remote courses")
		log.Println("[Sync] Error getting remote courses", wrappedErr)
		return wrappedErr
	}

	var syncLogs []models.LocalUpdate
	if err := db.Find(&syncLogs).Error; err != nil {
		wrappedErr := errors.HandleDBReadError(err)
		log.Println("[Sync] Error getting local updates", wrappedErr)
		return wrappedErr
	}

	for _, syncLog := range syncLogs {
		switch syncLog.Entity {
		case models.EntityAssignment:
			err := SyncAssignment(syncLog, remoteAssignments, db)
			if err != nil {
				log.Println("[Sync] Error syncing assignment:", err)
				return err
			}
		case models.EntityCourse:
			err := SyncCourse(syncLog, remoteCourses, db)
			if err != nil {
				log.Println("[Sync] Error syncing course:", err)
				return err
			}
		}
	}

	return nil
}

func SyncAssignment(syncLog models.LocalUpdate, remoteAssignments []map[string]string, db *gorm.DB) error {

	switch syncLog.Action {
	case "create":
		var localAssignment assignment.LocalAssignment
		if err := db.Where("id = ?", syncLog.EntityID).First(&localAssignment).Error; err != nil {
			return errors.HandleDBReadError(err)
		}

		remoteAssignment := &assignment.Assignment{
			LocalID:    localAssignment.ID,
			Title:      localAssignment.Title,
			Todo:       localAssignment.Todo,
			Deadline:   localAssignment.Deadline,
			CourseCode: localAssignment.CourseCode,
			TypeName:   localAssignment.TypeName,
			StatusName: localAssignment.StatusName,
			Priority:   localAssignment.Priority,
		}

		responseAssignment, err := client.CreateAssignment(remoteAssignment)
		if err != nil {
			return err
		}

		str_remote_id, ok := responseAssignment["id"].(string)
		if !ok {
			return errors.NewAppError(errors.SyncInvalidRemoteID, "Invalid remote assignment ID format", nil)
		}

		remote_id, err := strconv.Atoi(str_remote_id)
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to convert remote assignment ID to int")
		}

		localAssignment.RemoteID = uint(remote_id)

		if err := db.Save(localAssignment).Error; err != nil {
			return errors.HandleDBWriteError(err)
		}
	case "update", "delete":

		var localAssignment assignment.LocalAssignment
		if err := db.Unscoped().Where("id = ?", syncLog.EntityID).First(&localAssignment).Error; err != nil {
			return errors.HandleDBReadError(err)
		}

		remoteAssignment := findRemoteEntity(remoteAssignments, localAssignment.RemoteID)

		// If the remote assignment is not found, return an error
		if remoteAssignment == nil {
			return errors.NewAppError(errors.SyncRemoteNotFound, "Remote assignment not found", nil)
		}

		log.Println("[Sync] Syncing assignment", remoteAssignment["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))

		// Check if the remote assignment has been updated
		if remoteAssignment["updated_at"] < syncLog.UpdatedAt.Format(time.RFC3339) {

			remote_id_int := int(localAssignment.RemoteID)

			remote_id := strconv.Itoa(remote_id_int)

			if remote_id == "0" {
				return errors.NewAppError(errors.SyncInvalidRemoteID, "Remote assignment ID is 0", nil)
			}

			if err := client.UpdateAssignment(remote_id, syncLog.Column, syncLog.Value); err != nil {
				return err
			}

		} else {
			log.Println("Remote assignment has been updated", remoteAssignment["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
		}
	}

	return nil

}

func SyncCourse(syncLog models.LocalUpdate, remoteCourses []map[string]string, db *gorm.DB) error {

	switch syncLog.Action {
	case "create":
		var localCourse course.LocalCourse
		if err := db.Where("id = ?", syncLog.EntityID).First(&localCourse).Error; err != nil {
			return errors.HandleDBReadError(err)
		}
		remoteCourse := &course.Course{
			LocalID:         localCourse.ID,
			Name:            localCourse.Name,
			Code:            localCourse.Code,
			Color:           localCourse.Color,
			Semester:        localCourse.Semester,
			Schedule:        localCourse.Schedule,
			Credits:         localCourse.Credits,
			Location:        localCourse.Location,
			Instructor:      localCourse.Instructor,
			InstructorEmail: localCourse.InstructorEmail,
			StartDate:       localCourse.StartDate,
			EndDate:         localCourse.EndDate,
		}

		responseCourse, err := client.CreateCourse(remoteCourse)
		if err != nil {
			return err
		}

		str_remote_id, ok := responseCourse["id"].(string)
		if !ok {
			return errors.NewAppError(errors.SyncInvalidRemoteID, "Invalid remote course ID format", nil)
		}

		remote_id, err := strconv.Atoi(str_remote_id)
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to convert remote course ID to int")
		}

		localCourse.RemoteID = uint(remote_id)
		if err := db.Save(&localCourse).Error; err != nil {
			return errors.HandleDBWriteError(err)
		}

	case "update", "delete":

		var localCourse course.LocalCourse
		if err := db.Unscoped().Where("id = ?", syncLog.EntityID).First(&localCourse).Error; err != nil {
			return errors.HandleDBReadError(err)
		}
		remoteCourse := findRemoteEntity(remoteCourses, localCourse.RemoteID)

		// If the remote course is not found, return an error
		if remoteCourse == nil {
			return errors.NewAppError(errors.SyncRemoteNotFound, "Remote course not found", nil)
		}
		log.Println("[Sync] Syncing course", remoteCourse["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
		// Check if the remote course has been updated
		if remoteCourse["updated_at"] < syncLog.UpdatedAt.Format(time.RFC3339) {

			remote_id_int := int(localCourse.RemoteID)
			remote_id := strconv.Itoa(remote_id_int)

			if remote_id == "0" {
				return errors.NewAppError(errors.SyncInvalidRemoteID, "Remote course ID is 0", nil)
			}

			if err := client.UpdateCourse(remote_id, syncLog.Column, syncLog.Value); err != nil {
				return err
			}

		} else {
			log.Println("Remote course has been updated", remoteCourse["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
		}
	}

	return nil
}

func SyncNote(syncLog models.LocalUpdate, remoteNote map[string]string, db *gorm.DB) error {

	if remoteNote == nil {
		return errors.NewAppError(errors.SyncRemoteNotFound, "Remote note not found", nil)
	}

	log.Println("[Sync] Syncing note", remoteNote["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
	// Check if the remote course has been updated
	if remoteNote["updated_at"] < syncLog.UpdatedAt.Format(time.RFC3339) {

		sync_id_int := int(syncLog.EntityID)
		sync_id := strconv.Itoa(sync_id_int)
		if err := client.UpdateNote(sync_id, syncLog.Column, syncLog.Value); err != nil {
			return err
		}

	} else {
		log.Println("Remote note has been updated", remoteNote["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
	}

	return nil

}

func SyncUser(syncLog models.LocalUpdate) error {

	return client.UpdateUser(syncLog.Column, syncLog.Value)
}

func findRemoteEntity(remoteEntities []map[string]string, localEntityID uint) map[string]string {
	for _, remoteEntity := range remoteEntities {
		if remoteEntity["id"] == strconv.Itoa(int(localEntityID)) {
			return remoteEntity
		}
	}
	return nil
}
