package sync

import (
	"strconv"
	"time"
	"unipilot/internal/client"
	"unipilot/internal/errors"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"

	"gorm.io/gorm"
)

func MigrateAssignments(db *gorm.DB) error {

	count := 0

	remoteAssignments, err := client.GetAssignments()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote assignments for migration")
	}

	for _, ra := range remoteAssignments {

		localAssignment := assignment.LocalAssignment{
			RemoteID:   ra.ID,
			Title:      ra.Title,
			Todo:       ra.Todo,
			Deadline:   ra.Deadline,
			Link:       ra.Link,
			CourseCode: ra.CourseCode,
			TypeName:   ra.TypeName,
			StatusName: ra.StatusName,
			Priority:   ra.Priority,
		}

		if err := db.First(&localAssignment, "remote_id = ?", ra.ID).Error; err == nil {
			continue
		}

		if err := db.Create(&localAssignment).Error; err != nil {
			count++
			return errors.HandleDBCreateError(err)
		}
		count++
	}

	return nil
}

func MigrateCourses(db *gorm.DB) error {

	count := 0

	remoteCourses, err := client.GetCourses()
	if err != nil {
		return errors.Wrap(err, errors.SyncFailed, "Failed to get remote courses for migration")
	}

	for _, rc := range remoteCourses {
		remote_id, err := strconv.Atoi(rc["id"])
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to convert course remote ID to int")
		}

		start_date, err := time.Parse(time.DateOnly, rc["start_date"])
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to parse course start date")
		}

		end_date, err := time.Parse(time.DateOnly, rc["end_date"])
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to parse course end date")
		}

		credits, err := strconv.Atoi(rc["credits"])
		if err != nil {
			return errors.Wrap(err, errors.SyncDataConversionError, "Failed to convert course credits to int")
		}

		localCourse := course.LocalCourse{
			RemoteID:        uint(remote_id),
			Code:            rc["code"],
			Name:            rc["name"],
			Color:           rc["color"],
			Location:        rc["location"],
			StartDate:       start_date,
			EndDate:         end_date,
			Schedule:        rc["schedule"],
			Credits:         credits,
			Semester:        rc["semester"],
			Instructor:      rc["instructor"],
			InstructorEmail: rc["instructor_email"],
		}

		if err := db.First(&localCourse, "remote_id = ?", remote_id).Error; err == nil {
			continue
		}

		if err := db.Create(&localCourse).Error; err != nil {
			count++
			return errors.HandleDBCreateError(err)
		}
		count++
	}

	return nil
}
