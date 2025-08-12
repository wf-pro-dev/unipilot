package sync

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/models"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

func Sync(db *gorm.DB) error {

	// Get remote assignments
	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		log.Println("[Sync] Error getting remote assignments", err)
		return err
	}

	// Get remote courses
	remoteCourses, err := client.GetCourses()
	if err != nil {
		log.Println("[Sync] Error getting remote courses", err)
		return err
	}

	var syncLogs []models.LocalUpdate
	if err := db.Where("synced = ?", false).Find(&syncLogs).Error; err != nil {
		log.Println("[Sync] Error getting local updates", err)
		return err
	}

	for _, syncLog := range syncLogs {
		switch syncLog.Entity {
		case models.EntityAssignment:
			err := SyncAssignment(syncLog, findRemoteEntity(remoteAssignments, syncLog.EntityID), db)
			if err != nil {
				log.Println("[Sync] Error syncing assignment", err)
				return err
			}
		case models.EntityCourse:
			err := SyncCourse(syncLog, findRemoteEntity(remoteCourses, syncLog.EntityID), db)
			if err != nil {
				log.Println("[Sync] Error syncing course", err)
				return err
			}
		}
	}

	return nil
}

func SyncAssignment(syncLog models.LocalUpdate, remoteAssignment map[string]string, db *gorm.DB) error {

	// If the remote assignment is not found, return an error
	if remoteAssignment == nil {
		return fmt.Errorf("remote assignment not found")
	}

	log.Println("[Sync] Syncing assignment", remoteAssignment["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))

	// Check if the remote assignment has been updated
	if remoteAssignment["updated_at"] < syncLog.UpdatedAt.Format(time.RFC3339) {
		switch syncLog.Action {
		case "create":
			var localAssignment assignment.LocalAssignment
			if err := db.Where("id = ?", syncLog.EntityID).First(&localAssignment).Error; err != nil {
				return err
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

			localAssignment.RemoteID = uint(responseAssignment["id"].(float64))
			if err := db.Save(&localAssignment).Error; err != nil {
				return err
			}

		case "update":

			sync_id_int := int(syncLog.EntityID)

			sync_id := strconv.Itoa(sync_id_int)

			if err := client.SendAssignmentUpdate(sync_id, syncLog.Column, syncLog.Value); err != nil {
				return err
			}
		}

	} else {
		log.Println("Remote assignment has been updated", remoteAssignment["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
	}

	syncLog.SyncedAt = time.Now()
	syncLog.Synced = true
	if err := db.Save(&syncLog).Error; err != nil {
		return err
	}
	return nil

}

func SyncCourse(syncLog models.LocalUpdate, remoteCourse map[string]string, db *gorm.DB) error {

	// If the remote course is not found, return an error
	if remoteCourse == nil {
		return fmt.Errorf("remote course not found")
	}
	log.Println("[Sync] Syncing course", remoteCourse["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
	// Check if the remote course has been updated
	if remoteCourse["updated_at"] < syncLog.UpdatedAt.Format(time.RFC3339) {
		switch syncLog.Action {
		case "create":
			var localCourse course.LocalCourse
			if err := db.Where("id = ?", syncLog.EntityID).First(&localCourse).Error; err != nil {
				return err
			}
			remoteCourse := &course.Course{
				LocalID:         localCourse.ID,
				Name:            localCourse.Name,
				Code:            localCourse.Code,
				Color:           localCourse.Color,
				Semester:        localCourse.Semester,
				Schedule:        localCourse.Schedule,
				Credits:         localCourse.Credits,
				RoomNumber:      localCourse.RoomNumber,
				Instructor:      localCourse.Instructor,
				InstructorEmail: localCourse.InstructorEmail,
				StartDate:       localCourse.StartDate,
				EndDate:         localCourse.EndDate,
			}

			responseCourse, err := client.CreateCourse(remoteCourse)
			if err != nil {
				return err
			}

			localCourse.RemoteID = uint(responseCourse["id"].(float64))
			if err := db.Save(&localCourse).Error; err != nil {
				return err
			}
		case "update":
			sync_id_int := int(syncLog.EntityID)
			sync_id := strconv.Itoa(sync_id_int)
			if err := client.SendCourseUpdate(sync_id, syncLog.Column, syncLog.Value); err != nil {
				return err
			}
		}
	} else {
		log.Println("Remote course has been updated", remoteCourse["updated_at"], syncLog.UpdatedAt.Format(time.RFC3339))
	}

	syncLog.SyncedAt = time.Now()
	syncLog.Synced = true
	if err := db.Save(&syncLog).Error; err != nil {
		return err
	}
	return nil
}

func findRemoteEntity(remoteEntities []map[string]string, localEntityID uint) map[string]string {
	for _, remoteEntity := range remoteEntities {
		if remoteEntity["local_id"] == strconv.Itoa(int(localEntityID)) {
			return remoteEntity
		}
	}
	return nil
}
