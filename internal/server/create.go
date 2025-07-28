package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/types"

	"gorm.io/gorm"
)

func WebhookCreateHandler(w http.ResponseWriter, r *http.Request, payload types.NotionWebhookPayload, u *user.User) {

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

	// 1. Get the page id
	page_id := payload.Entity.Id

	// 2. Get the page properties
	new_page, err := assignment.GetPage(page_id, u.NotionAPIKey)
	if err != nil {
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error getting page properties: %s", err))
		return
	}

	var pageResp types.PageRequest
	if err := json.Unmarshal(new_page, &pageResp); err != nil {
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error unmarshalling page: %s", err))
		return
	}

	var properties = pageResp.Properties

	course_notion_id := properties.Courses.Relation[0].ID
	course_code := course.Get_Course_byNotionID(course_notion_id, db).Code

	deadline, err := time.Parse(time.DateOnly, properties.Deadline.Date.Start)
	if err != nil {

		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error parsing deadline: %s", err))
		return
	}

	PrintLog(fmt.Sprintf("User id : %d,\n page_id : %s\n ", u.ID, page_id))

	a := assignment.Assignment{
		UserID:     u.ID,
		CourseCode: course_code,
		Title:      properties.AssignmentName.Title[0].PlainText,
		TypeName:   properties.Type.Select["name"],
		Deadline:   deadline,
		Todo:       properties.TODO.RichText[0].PlainText,
		StatusName: properties.Status.Status.Name,
		Link:       properties.Link.URL,
		NotionID:   page_id}

	result := tx.Create(&a)
	if result.Error != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusConflict, fmt.Sprintf("Error creating assignment in database", err))
		return
	}

	tx.Commit()

	a_map := a.ToMap()

	a_map["deadline"] = a.Deadline.Format(time.RFC3339)

	a_map["created_at"] = a.CreatedAt.Format(time.RFC3339)

	a_map["updated_at"] = a.UpdatedAt.Format(time.RFC3339)

	PrintLog(fmt.Sprintf("Assignment created: %v by user %v", a_map["title"], u.ID))

	if sseServer != nil {
		sseServer.SendNotification(
			u.ID,
			"create",
			"assignment",
			a.NotionID,
			fmt.Sprintf("New assignment created: %s", a.Title),
			a_map,
		)

	}
}
