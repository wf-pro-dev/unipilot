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
	"unipilot/internal/services/markdown"

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
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Error getting notes for user id = %d : %s", userID, err))
		return
	}

	var notesMap []map[string]string
	for _, n := range notes {
		notesMap = append(notesMap, n.ToMap())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User's notes retrieved successfully",
		"notes":   notesMap,
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
		LocalID    string `json:"local_id"`
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

	// Generate content and keywords using Gemini
	geminiRequest := &gemini.GeminiRequest{
		Title:      input.Title,
		Subject:    input.Subject,
		CourseName: input.CourseCode,
	}

	geminiResponse, err := gemini.GenerateNote(geminiRequest)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate note content: %v", err))
		return
	}

	// Parse markdown content to HTML for storage
	markdownService := markdown.NewMarkdownService()
	htmlContent, err := markdownService.ParseToHTML(geminiResponse["content"])
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse markdown content: %v", err))
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
		CourseCode: input.CourseCode,
		Title:      input.Title,
		Subject:    input.Subject,
		Keywords:   geminiResponse["keywords"],
		Content:    htmlContent,
	}

	result := tx.Create(&nVal)
	if result.Error != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating note in database: %v", result.Error))

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
	noteMap := nVal.ToMap()
	if noteMap == nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError, "Failed to process note data")
		return
	}

	tx.Commit()

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Note created successfully",
		"note":    noteMap,
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
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to convert note ID to int: %s", err))
		return
	}


	var n note.Note
	if err := tx.Where("id = ? AND user_id = ?", uint(int_id), userID).First(&n).Error; err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to get note: %s", err))
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE notes SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), n.ID).Error; err != nil {

		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error updating note in database: %s", err))
		return
	}

	PrintLog(fmt.Sprintf("user_id %d column %s value %s",
		userID, updateData.Column, updateData.Value))

	tx.Commit()

}
