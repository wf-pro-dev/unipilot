package server

import (
	"fmt"
	"net/http"
	"time"

	"unipilot/internal/models/assignment"
	"unipilot/internal/types"

	"gorm.io/gorm"
)

func WebhookDeleteHandler(w http.ResponseWriter, r *http.Request, payload types.NotionWebhookPayload) {

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

	a, err := assignment.Get_Assignment_byNotionID(payload.Entity.Id, tx)

	if err != nil {
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error getting assignment: %s", err))
		return
	}

	a_map := a.ToMap()

	err = tx.Delete(&a).Error
	if err != nil {
		tx.Rollback()
		PrintERROR(w, http.StatusInternalServerError,
			fmt.Sprintf("Error deleting assignment: %s", err))
		return
	}

	tx.Commit()

	PrintLog(fmt.Sprintf("Assignment deleted: %v by user %v", a_map["title"], a.UserID))

	a_map["deadline"] = a.Deadline.Format(time.RFC3339)

	a_map["deadline"] = a.CreatedAt.Format(time.RFC3339)

	a_map["deadline"] = a.UpdatedAt.Format(time.RFC3339)

	if sseServer != nil {
		sseServer.SendNotification(
			a.UserID,
			"delete",
			"assignment",
			payload.Entity.Id,
			fmt.Sprintf("Assignment deleted: %s", a.Title),
			a_map,
		)
	}
}
