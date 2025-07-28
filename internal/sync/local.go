package sync

import (
	"fmt"
	"strconv"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

func MigrateAssignments(db *gorm.DB) error {

	count := 0

	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return err
	}

	for _, ra := range remoteAssignments {

		deadline, err := time.Parse(time.DateOnly, ra["deadline"])
		if err != nil {

			return fmt.Errorf("Error formating deadline : %s", err)
		}

		remote_id, err := strconv.Atoi(ra["id"])
		if err != nil {
			return fmt.Errorf("Error formating remote_id : %s", err)
		}

		localAssignment := assignment.LocalAssignment{
			RemoteID:   uint(remote_id),
			Title:      ra["title"],
			Todo:       ra["todo"],
			Deadline:   deadline,
			Link:       ra["link"],
			CourseCode: ra["course_code"],
			TypeName:   ra["type"],
			StatusName: ra["status"],
			NotionID:   ra["notion_id"],
			SyncStatus: assignment.SyncStatusSynced,
		}

		if err := db.First(&localAssignment, "remote_id = ?", remote_id).Error; err == nil {
			continue
		}

		if err := db.Create(&localAssignment).Error; err != nil {
			count++
			return err
		}
		count++
	}

	return nil
}

func MigrateCourses(db *gorm.DB) error {

	count := 0

	remoteCourses, err := client.GetCourses()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return err
	}

	for _, rc := range remoteCourses {
		remote_id, err := strconv.Atoi(rc["id"])
		if err != nil {
			return fmt.Errorf("Error formating remote_id : %s", err)
		}
		localCourse := course.LocalCourse{
			RemoteID:   uint(remote_id),
			Code:       rc["code"],
			Name:       rc["name"],
			NotionID:   rc["notion_id"],
			Duration:   rc["duration"],
			RoomNumber: rc["room_number"],
			SyncStatus: course.SyncStatusSynced,
		}

		if err := db.First(&localCourse, "remote_id = ?", remote_id).Error; err == nil {
			continue
		}

		if err := db.Create(&localCourse).Error; err != nil {
			count++
			return err
		}
		count++
	}

	return nil
}
