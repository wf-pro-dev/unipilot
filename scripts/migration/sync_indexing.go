package main

import (
	"fmt"
	"strconv"
	"unipilot/internal/client"
	"unipilot/internal/models"
	"unipilot/internal/services/utils"
)

func sync_indexing_assignments() {

	// Get the local database
	db, err := utils.GetUserDB()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Get all local assignment from local database
	assignments := []models.LocalAssignment{}
	err = db.Find(&assignments).Error
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Upadete remote assignment local id with the local assignment id
	for _, assignment := range assignments {
		assignment_id := strconv.Itoa(int(assignment.ID))
		remote_id := strconv.Itoa(int(assignment.RemoteID))

		if err := client.UpdateAssignment(remote_id, "local_id", assignment_id); err != nil {
			fmt.Printf("ERROR : %s", err)
		} else {
			fmt.Printf("Updated assignment %s\n", assignment_id)
		}
	}

}

func sync_indexing_courses() {

	// Get the local database
	db, err := utils.GetUserDB()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Get all local assignment from local database
	courses := []models.LocalCourse{}
	err = db.Find(&courses).Error
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Upadete remote assignment local id with the local assignment id
	for _, course := range courses {

		course_id := strconv.Itoa(int(course.ID))
		remote_id := strconv.Itoa(int(course.RemoteID))

		if err := client.UpdateCourse(remote_id, "local_id", course_id); err != nil {
			fmt.Printf("ERROR : %s", err)
		} else {
			fmt.Printf("Updated course %s\n", course_id)
		}
	}

}

// func main() {
// 	sync_indexing_assignments()
// 	sync_indexing_courses()
// }
