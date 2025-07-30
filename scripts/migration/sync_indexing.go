package main

import (
	"fmt"
	"strconv"
	"unipilot/internal/client"
	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/storage"
)


func sync_indexing_assignments() {

	// Get the local database
	db, _, err := storage.GetLocalDB()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Get all local assignment from local database
	assignments := []assignment.LocalAssignment{}
	err = db.Find(&assignments).Error
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Upadete remote assignment local id with the local assignment id
	for _, assignment := range assignments {
		assignment_id_int := int(assignment.RemoteID)
		assignment_id := strconv.Itoa(assignment_id_int)

		if err := client.SendAssignmentUpdate(assignment_id, "local_id", assignment_id); err != nil {
			fmt.Printf("ERROR : %s", err)
		} else {
			fmt.Printf("Updated assignment %s\n", assignment_id)
		}
	}

}

func sync_indexing_courses() {

	// Get the local database
	db, _, err := storage.GetLocalDB()
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Get all local assignment from local database
	courses := []course.LocalCourse{}
	err = db.Find(&courses).Error
	if err != nil {
		fmt.Printf("ERROR : %s", err)
		return
	}

	// Upadete remote assignment local id with the local assignment id
	for _, course := range courses {
		course_id_int := int(course.RemoteID)
		course_id := strconv.Itoa(course_id_int)

		if err := client.SendCourseUpdate(course_id, "local_id", course_id); err != nil {
			fmt.Printf("ERROR : %s", err)
		} else {
			fmt.Printf("Updated course %s\n", course_id)
		}
	}

}

func main() {
	sync_indexing_assignments()
	sync_indexing_courses()
}