package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"

	"unipilot/internal/models/note"
	"unipilot/internal/models/user"
	"unipilot/internal/server"
	"unipilot/internal/services/gemini"
)

func GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	var notes []note.Note
	if err := db.Where("user_id = ?", userID).Find(&notes).Error; err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting notes from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
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

	server.LogInfo("Notes retrieved successfully",
		"request_id", requestID,
		"user_id", userID,
		"count", len(notesMap),
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"NOTES", "READ"},
	)
}

func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

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
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "REQUEST"},
		)
		return
	}

	// Validate all required fields
	if input.CourseCode == "" || input.Title == "" || input.Subject == "" {
		err := fmt.Errorf("missing required fields: course code: %s, title: %s, subject: %s", input.CourseCode, input.Title, input.Subject)
		server.ResponseError(w, err, http.StatusBadRequest, "Missing required fields",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "MISSING_REQUIRED_FIELDS"},
		)
		return
	}

	keywords := input.Keywords
	content := input.Content

	//server.PrintLOG([]string{"INFO", "CREATE", "NOTE"}, fmt.Sprintf("Keywords: %s, Content: %s", keywords, content))

	// Gnerate note gemini data if missing
	if keywords == "" && content == "" {

		// Generate content and keywords using Gemini
		geminiRequest := &gemini.GeminiRequest{
			Title:      input.Title,
			Subject:    input.Subject,
			CourseName: input.CourseCode,
		}

		geminiResponse, err := gemini.GenerateNote(geminiRequest)
		if err != nil {
			server.ResponseError(w, err, http.StatusInternalServerError, "Error generating note content with Gemini",
				"request_id", requestID,
				"user_id", userID,
				"duration", time.Since(startTime).Milliseconds(),
				"tags", []string{"NOTES", "GEMINI"},
			)
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
		Videos:     input.Videos,
		Keywords:   keywords,
		Content:    content,
	}

	result := tx.Create(&nVal)
	if result.Error != nil {
		tx.Rollback()
		server.ResponseError(w, result.Error, http.StatusConflict, "Error creating note in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
		return
	}

	// Convert to map safely
	noteMap := nVal.ToMap()
	if noteMap == nil {
		tx.Rollback()
		server.ResponseError(w, fmt.Errorf("failed to process note data"), http.StatusInternalServerError, "Error processing note data",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "MARSHALLING"},
		)
		return
	}

	tx.Commit()

	// gRPC -> c : TO BE MOVE IN DOCKER

	// Send a notification to all the users linked

	newN, err := note.Get_Note_byID(nVal.ID, userID, db)
	if err != nil {
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting note from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
		return
	}

	// nJson, err := json.Marshal(newN)
	_, err = json.Marshal(newN)
	if err != nil {
		server.LogWarn(
			"Error marshalling notification", err,
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "MARSHALLING"},
		)
	}

	// link_users, err := newN.Course.GetLinkUsers(db)
	_, err = newN.Course.GetLinkUsers(db)
	if err != nil {
		server.LogWarn(
			"Error getting users linked to course", err,
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
	}

	/*if sseServer != nil {
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
	}*/

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Note created successfully",
		"note":    noteMap,
	})

	server.LogInfo("Note created successfully",
		"request_id", requestID,
		"user_id", userID,
		"note_id", nVal.ID,
		"title", nVal.Title,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"NOTES", "WRITE"},
	)
}

func UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	startTime := r.Context().Value("start_time").(time.Time)
	requestID := r.Context().Value("request_id").(string)
	currentUser := r.Context().Value("user").(user.User)
	db := r.Context().Value("db").(*gorm.DB)
	userID := currentUser.ID

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var updateData struct {
		ID     string `json:"id"`
		Value  string `json:"value"`
		Column string `json:"column"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Invalid request body",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "REQUEST"},
		)
		return
	}

	int_id, err := strconv.Atoi(updateData.ID)
	if err != nil {
		server.ResponseError(w, err, http.StatusBadRequest, "Error converting note ID to int",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "INVALID_NOTE_ID"},
		)
		return
	}

	n, err := note.Get_Note_byID(uint(int_id), userID, tx)
	if err != nil {
		tx.Rollback()
		server.ResponseError(w, err, http.StatusInternalServerError, "Error getting note from database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
		return
	}

	if err := tx.Exec(fmt.Sprintf("UPDATE notes SET %s = ?, updated_at = ? WHERE id = ?", updateData.Column),
		updateData.Value, time.Now().Format(time.RFC3339), n.ID).Error; err != nil {
		tx.Rollback()
		server.ResponseError(w, err, http.StatusInternalServerError, "Error updating note in database",
			"request_id", requestID,
			"user_id", userID,
			"duration", time.Since(startTime).Milliseconds(),
			"tags", []string{"NOTES", "DB"},
		)
		return
	}

	tx.Commit()

	server.LogInfo("Note updated successfully",
		"request_id", requestID,
		"user_id", userID,
		"note_id", n.ID,
		"update", updateData,
		"duration", time.Since(startTime).Milliseconds(),
		"tags", []string{"NOTES", "WRITE"},
	)
}
