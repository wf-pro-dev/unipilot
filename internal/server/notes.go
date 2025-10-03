package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models"
	"unipilot/internal/models/note"
	"unipilot/internal/models/notifications"
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
		UserID     string `json:"user_id"`
		CourseCode string `json:"course_code"`
		Title      string `json:"title"`
		Subject    string `json:"subject"`
		Content    string `json:"content"`
		Keywords   string `json:"keywords"`
		Videos     string `json:"videos"`

	}
	

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		PrintERROR(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	

	// Validate all required fields
	if input.CourseCode == "" || input.Title == "" || input.Subject == "" {
		PrintLog(fmt.Sprintf("Course code : %s, title: %s, subject: %s",input.CourseCode,input.Title,input.Subject))
		PrintERROR(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	keywords := input.Keywords
	content :=  input.Content
	
	// Gnerate note gemini data if missing
	if keywords == "" && content== "" {


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

		keywords = geminiResponse.Keywords
		content = geminiResponse.Content
	}

	

	nVal := note.Note{
		UserID:     userID,
		CourseCode: input.CourseCode,
		Title:      input.Title,
		Subject:    input.Subject,
		Videos:	    input.Videos,
		Keywords:   keywords,
		Content:    content,
	}

	result := tx.Create(&nVal)
	if result.Error != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating note in database: %v", result.Error))
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

	db = db.Debug()
	
	// Send a notification to all the users linked
	
	newN, err := note.Get_Note_byID(nVal.ID, userID ,db)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to get note: %s", err))
		return
	}
	
	nJson, err := json.Marshal(newN)
	if err != nil {
		log.Printf("[Error] error marshalling notification : %v ",err)
	}


	link_users, err :=  newN.Course.GetLinkUsers(db)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("failed to getting users link to course assignment: %s", err))
		return
	}

	
	
	PrintLog(fmt.Sprintf("link users : %v, sseServer : %v", link_users, sseServer != nil))
	if sseServer != nil {
		// 2. Send link info to users via SSE (field data)
		for _,sendeeID := range link_users {
			if sendeeID != userID {
				PrintLog(fmt.Sprintf("sending to : %d ",sendeeID))
				sseServer.SendNotification(
					uint(sendeeID),
					userID,
					models.EntityNote,
					newN.Course.ID,
					notifications.NotificationNoteUpdate,
					newN.Title,
					fmt.Sprintf("%s shared a new note on %s", newN.User.Username, newN.CourseCode),
					"note",
					string(nJson),

				)
			}
		}
	}

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


	n, err := note.Get_Note_byID(uint(int_id), userID, db) 
	if err != nil {
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
