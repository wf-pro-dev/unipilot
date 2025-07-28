package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"unipilot/internal/models/assignment"
	"unipilot/internal/models/course"
	"unipilot/internal/models/user"
	"unipilot/internal/types"

	"gorm.io/gorm"
)

func WebhookUpdateHandler(w http.ResponseWriter, r *http.Request, payload types.NotionWebhookPayload, u *user.User) {

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

	//  Get the page id
	page_id := payload.Entity.Id

	// 6. Loop through the properties
	for _, property_id := range payload.Data.Properties {

		// 7. Get the updated property
		property, err := assignment.GetPageProperties(page_id, property_id, u.NotionAPIKey)
		if err != nil {
			PrintERROR(w, http.StatusInternalServerError, fmt.Sprintf("Fetching properties: %s", err))
			return
		}

		// For debugging
		// var property_map map[string]interface{}
		// json.Unmarshal(property, &property_map)

		// 8. Get the column name from the property id
		column := types.COLUMNS[property_id]
		if column == "" {
			PrintERROR(w, http.StatusInternalServerError,
				fmt.Sprintf("Column not found for id: %s", property_id))
			return
		}

		// 9. Get the new value from the property
		value := GetValue(w, property, column, tx)

		// Log the update
		PrintLog(fmt.Sprintf("page_id %s property_id %s column %s value %s",
			page_id, property_id, column, value))

		if value != "" {

			// Update the assignment in the database

			a, err := assignment.Get_Assignment_byNotionID(page_id, tx)
			if err != nil {
				PrintERROR(w, http.StatusInternalServerError,
					fmt.Sprintf("Error getting assignment: %s", err))
				return
			}

			if err := tx.Exec(fmt.Sprintf("UPDATE assignments SET %s = ?, updated_at = ? WHERE id = ?", column),
				value, time.Now().Format(time.RFC3339), a.ID).Error; err != nil {

				tx.Rollback()
				PrintERROR(w, http.StatusInternalServerError,
					fmt.Sprintf("Error updating assignment in database: %s", err))
				return
			}

			if sseServer != nil {
				sseServer.SendNotification(
					u.ID,
					"update",
					"assignment",
					a.NotionID,
					fmt.Sprintf("Assignment updated: %s %s is now %s", a.Title, column, value),
					map[string]string{
						"id":     strconv.Itoa(int(a.ID)),
						"column": column,
						"value":  value,
					},
				)
			} else {
				PrintLog("sseServer is nil\n")
			}

		}
	}
	tx.Commit()

}

func GetValue(w http.ResponseWriter, property []byte, column string, db *gorm.DB) string {
	var value string
	switch column {

	case "course_code":
		var coursesType struct {
			Courses []struct {
				Relation struct {
					ID string `json:"id"`
				} `json:"relation"`
			} `json:"results"`
		}
		json.Unmarshal(property, &coursesType)

		if len(coursesType.Courses) > 0 {
			course := course.Get_Course_byNotionID(coursesType.Courses[0].Relation.ID, db)

			if course == nil {
				value = ""
				err := fmt.Errorf("course not found")
				PrintERROR(w, http.StatusInternalServerError,
					fmt.Sprintf("Error getting course: %s", err))
			} else {
				value = course.Code
			}
		}

	case "deadline":
		var dateType struct {
			Date struct {
				Start string `json:"start"`
			} `json:"date"`
		}

		json.Unmarshal(property, &dateType)
		value = dateType.Date.Start

	case "link":
		var linkType struct {
			URL string `json:"url"`
		}
		json.Unmarshal(property, &linkType)
		value = linkType.URL

	case "todo":
		var todoType struct {
			Results []struct {
				RichText struct {
					PlainText string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"results"`
		}

		json.Unmarshal(property, &todoType)
		value = todoType.Results[0].RichText.PlainText

	case "title":
		var titleType struct {
			Results []struct {
				Title struct {
					PlainText string `json:"plain_text"`
				} `json:"title"`
			} `json:"results"`
		}
		json.Unmarshal(property, &titleType)
		value = titleType.Results[0].Title.PlainText

	case "type_name":

		var selectType struct {
			Select map[string]string `json:"select"`
		}

		json.Unmarshal(property, &selectType)
		value = selectType.Select["name"]

	case "status_name":
		var statusType struct {
			Status struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"status"`
		}
		json.Unmarshal(property, &statusType)
		value = statusType.Status.Name
	default:
		value = ""
	}

	return value

}
