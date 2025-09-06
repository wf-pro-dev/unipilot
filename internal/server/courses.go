package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"log"
	
	"unipilot/internal/models"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/models/notifications"

	"gorm.io/gorm"
)

func GetCourseHandler(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	var courses []course.Course
	if err := db.Where("user_id = ?", userID).Find(&courses).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", userID, err))
		return
	}

	var coursesMap []map[string]string
	for _, a := range courses {
		coursesMap = append(coursesMap, a.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's Assignments retrieved successfully",
		"courses": coursesMap,
	})
}

func CreateCourseHandler(w http.ResponseWriter, r *http.Request) {

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var input struct {
		LocalID         string `json:"local_id"`
		Name            string `json:"name"`
		Code            string `json:"code"`
		Color           string `json:"color"`
		Semester        string `json:"semester"`
		Schedule        string `json:"schedule"`
		Credits         string `json:"credits"`
		Location        string `json:"location"`
		StartDate       string `json:"start_date"`
		EndDate         string `json:"end_date"`
		Instructor      string `json:"instructor"`
		InstructorEmail string `json:"instructor_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	PrintLog(fmt.Sprintf("course input data local ID : %s Title: %s : Course code: %s Type : %s Deadline : %s \n", input.LocalID, input.Instructor, input.EndDate, input.Semester, input.StartDate))
	PrintLog(fmt.Sprintf("course input data local ID : %s \n", input.Code))

	// Validate all required fields
	if input.LocalID == "" || input.Code == "" || input.Semester == "" || input.Instructor == "" || input.StartDate == "" || input.EndDate == "" {
		PrintERROR(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	start_date, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, "Invalid start date format")
		return
	}

	end_date, err := time.Parse(time.DateOnly, input.EndDate)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, "Invalid start date format")
		return
	}

	credits, err := strconv.Atoi(input.Credits)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating credits : %s", err))

		return
	}
	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating local_id : %s", err))

		return
	}

	cVal := course.Course{
		UserID:          userID,
		LocalID:         uint(local_id),
		Name:            input.Name,
		Code:            input.Code,
		Color:           input.Color,
		Semester:        input.Semester,
		Schedule:        input.Schedule,
		Credits:         credits,
		Location:        input.Location,
		StartDate:       start_date,
		EndDate:         end_date,
		Instructor:      input.Instructor,
		InstructorEmail: input.InstructorEmail,
	}

	result := tx.Create(&cVal)
	if result.Error != nil {
		PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating assignment in database", err))
		return
	}

	cObj := &cVal

	c, err := course.Get_Course_byId(cObj.ID, tx)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting course: %s", err))
		return
	}
	/*notion_id, err := a.Add_Notion()
	if err != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error creating assignment in notion", err))
		return
	}

	a.NotionID = notion_id
	err = tx.Save(&a).Error
	if err != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error updating new assignment", err))
		return
	}*/

	// Convert to map safely
	courseMap := c.ToMap()
	if courseMap == nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, "Failed to process course data")
		return
	}

	tx.Commit()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Assignment created successfully",
		"course":  courseMap,
	})

}
func UpdateCourseHandler(w http.ResponseWriter, r *http.Request) {

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	/*userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}*/

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updateData struct {
		ID     string `json:"id"`
		Value  string `json:"value"`
		Column string `json:"column`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}

	int_id, err := strconv.Atoi(updateData.ID)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to convert assignment ID to int: %s", err))
		return
	}

	c, err := course.Get_Course_byId(uint(int_id), tx)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting course: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE courses SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), c.ID).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}

	/*value := updateData.Value

	//PrintLog(fmt.Sprintf("column : %s, value :%s, user id:%s", updateData.Column, value, dbVal.(string) ))
	if updateData.Column == "course_code" {
		//PrintLog(fmt.Sprintf("column : %s, value :%s, user id:%s", updateData.Column, value, dbVal.(string) ))

		c, err := course.Get_Course_byCode(value, strconv.Itoa(int(userID)), tx)
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting new course: %s", err))
			return
		}
		PrintLog(fmt.Sprintf("Course %s", c.ToMap()))

		value = c.NotionID
		PrintLog(fmt.Sprintf("Course notion id: %s", value))
	}

	var obj map[string]string

	if updateData.Column == "status_name" {
		var status = models.Get_AssignmentStatus_byName(value, tx)
		obj = status.ToMap()
	} else if updateData.Column == "type_name" {
		var t = models.Get_AssignmentType_byName(value, tx)
		obj = t.ToMap()
	}

	err = a.Update_Notion(updateData.Column, value, obj)
	if err != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error updating assignment in notion", err))
		return
	}*/

	PrintLog(fmt.Sprintf("user_id %s course %d column %s value %s",
		userIDVal, c.ID, updateData.Column, updateData.Value))

	tx.Commit()

}

func LinkRequestCourseHandler(w http.ResponseWriter, r *http.Request){

	dbVal := r.Context().Value("db")
	if dbVal == nil {
		PrintERROR(w, http.StatusInternalServerError, "Database connection not found")
		return
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		PrintERROR(w, http.StatusInternalServerError, "Invalid database connection")
		return
	}

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		PrintERROR(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		PrintERROR(w, http.StatusUnauthorized, "Invalid user ID format")
		return
	}

	var currentUser user.User
	if err := db.First(&currentUser, userID).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, "Database error")
		return
	}
	
	var linkRequestData struct {
		CourseCode string `json:"course_code"`
		UsersID []uint `json:"users_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&linkRequestData)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body %s", err))
		return
	}
	
	// 1. Get send course informations
	c, err := course.Get_Course_byCode(linkRequestData.CourseCode, userID, db)

	cJson, err := json.Marshal(c)
	if err != nil {
		log.Printf("[Error] error marshalling notification : %v ",err)
		
	}
	if sseServer != nil {

		// 2. Send link info to users via SSE (field data)
		for _,sendeeID := range linkRequestData.UsersID { 
			sseServer.SendNotification(
				uint(sendeeID),
				userID,
				models.EntityCourse,
				c.ID,
				notifications.NotificationSync,
				c.Name,
				fmt.Sprintf("%s shared a course with you : %s", currentUser.Username , c.Code),
				"sync",
				string(cJson),

			)
			PrintLog(fmt.Sprintf("Sending sendee : %v, course name : %s", uint(sendeeID), c.Name))
		}
	}
	// Infos : Course All except user_id, Sender (name)

	PrintLog(fmt.Sprintf("Course ID : %v, Users ID : %v",linkRequestData.CourseCode, linkRequestData.UsersID))
	
}
