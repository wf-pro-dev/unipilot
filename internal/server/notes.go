package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models/note"
	"unipilot/internal/models/course"
	"unipilot/internal/services/gemini"
	
	"gorm.io/gorm"
)


func GetNoteHandler(w http.ResponseWriter, r *http.Request) {
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

	var notes []note.Note
	if err := db.Where("user_id = ?", userID).Find(&notes).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting assignment for user id = %d : %s", userID, err))
		return
	}

	var notesMap []map[string]string
	for _, n := range notes {
		notesMap = append(notesMap, n.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "User's notes retrieved successfully",
		"notes": notesMap,
	})

}
func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {

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
		LocalID	   string `json:"local_id"`	
		UserID     string `json:"user_id"`
		CourseCode string `json:"course_code"`
		Title      string `json:"title"`
		Subject    string `json:"subject"`
	}
	

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	


	// Validate all required fields
	if input.LocalID == "" || input.CourseCode == "" || input.Title == "" || input.Subject == "" {
		PrintERROR(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Generate content and keywords
	var request gemini.GeminiRequest
	
	var c course.Course
	if err := db.Where("code = ?",input.CourseCode).Find(&c).Error ; err != nil {
	
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Found no related course: %v", err))
		return
	}

	request.Title = input.Title
	request.Subject = input.Subject
	request.CourseName = c.Name
	

	response, err := gemini.GenerateNote(&request)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid gemini response: %v", err))
		return
	}


	PrintLog(response.Content)
	
	local_id, err := strconv.Atoi(input.LocalID)
	if err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Error formating local_id : %s", err))

		return 
	}
	

	nVal := note.Note{
		UserID:     userID,
		LocalID:    uint(local_id),
		CourseCode: input.CourseCode,
		Title:      input.Title,
		Subject:    input.Subject,
		Keywords:   response.Keywords,
		Content:    response.Content,
	}

	result := tx.Create(&nVal)
	if result.Error != nil {
		PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating assignment in database", err))
		return
	}

	nObj := &nVal

	n, err := note.Get_Note_byID(nObj.ID, userID, tx)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting assignment: %s", err))
		return
	}

	// Convert to map safely
	noteMap := n.ToMap()
	if noteMap == nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, "Failed to process assignment data")
		return
	}

	tx.Commit()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Assignment created successfully",
		"note": noteMap,
	})

}
func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {

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

	n, err := note.Get_Note_byLocalID(uint(int_id), userID, tx)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting assignment: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE notes SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),	
		updateData.Value, time.Now().Format(time.RFC3339), n.ID).Error; err != nil {

		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating assignment in database: %s", err))
		return
	}


	PrintLog(fmt.Sprintf("user_id %s column %s value %s",
		userIDVal, updateData.Column, updateData.Value))

	tx.Commit()

}

