package sync

import (
	Errors "errors"
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

		localCourse := course.LocalCourse{
			RemoteID:        rc.ID,
			Code:            rc.Code,
			Name:            rc.Name,
			Color:           rc.Color,
			Location:        rc.Location,
			StartDate:       rc.StartDate,
			EndDate:         rc.EndDate,
			Schedule:        rc.Schedule,
			Credits:         rc.Credits,
			Semester:        rc.Semester,
			Instructor:      rc.Instructor,
			InstructorEmail: rc.InstructorEmail,
		}

		if err := db.Create(&localCourse).Error; err != nil {
			if Errors.Is(err, gorm.ErrDuplicatedKey) {
				continue
			}
			count++
			return errors.HandleDBCreateError(err)
		}
		count++
	}

	return nil
}
