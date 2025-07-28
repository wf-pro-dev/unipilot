package assignment

import (
	"encoding/json"
	"fmt"

	"unipilot/internal/services/notion"
	"unipilot/internal/types"
)

func NewRichText(content string) types.RichText {
	return types.RichText{
		Type: "text",
		Text: &types.TextContent{
			Content: content,
			Link:    nil,
		},
		Annotations: &types.TextAnnotation{
			Bold:          false,
			Italic:        false,
			Strikethrough: false,
			Underline:     false,
			Code:          false,
			Color:         "default",
		},
		PlainText: content,
		Href:      nil,
	}
}

// AddAssignmentToNotion adds an assignment to Notion efficiently
func (a *Assignment) Add_Notion() (string, error) {

	assign := a.ToMap()

	// Create the request with strongly typed fields
	req := types.PageRequest{}
	req.Parent.Type = "database_id"
	req.Parent.DatabaseID = a.User.AssignmentsDbId
	// Set deadline
	req.Properties = &types.Properties{
		Deadline: types.Deadline{
			ID:   "_UjC",
			Type: "date",
			Date: &types.DateObject{
				Start: assign["deadline"], // 2025-06-05T00:00:00.000Z
			},
		},
		Courses: types.Courses{
			ID:   "w%3FC%3B",
			Type: "relation",
			Relation: []types.Relation{
				{
					ID: a.Course.NotionID,
				},
			},
		},
		Type: types.Type{
			ID:     "S~Ce",
			Type:   "select",
			Select: a.Type.ToMap(),
		},
		Status: types.Status{
			ID:   "%5Bm%5Cs",
			Type: "status",
			Status: &types.StatusObject{
				ID:    "3aa77cf8-c39e-4c7b-b7d2-ab15ae43ff23",
				Name:  "Not started",
				Color: "default",
			},
		},
		TODO: types.TODO{
			ID:   "%5DJfC",
			Type: "rich_text",
		},
		AssignmentName: types.AssignmentName{
			ID:   "title",
			Type: "title",
		},

		Link: types.Link{
			ID:   "jgPD",
			Type: "url",
			URL:  a.Link,
		},
	}
	// Set TODO
	todo_obj := types.TODO{
		ID:   "%5DJfC",
		Type: "rich_text",
	}
	todoText := NewRichText(assign["todo"])
	todo_obj.RichText = []types.RichText{todoText}
	req.Properties.TODO = todo_obj

	// Set title
	assignment_name_obj := types.AssignmentName{
		ID:   "title",
		Type: "title",
	}
	titleText := NewRichText(assign["title"])
	assignment_name_obj.Title = []types.RichText{titleText}
	req.Properties.AssignmentName = assignment_name_obj

	resp, err := notion.SendNotionRequest(req, "POST", "pages", a.User.NotionAPIKey)
	if err != nil {
		return "", err
	}

	// Parse response
	type NotionResponse struct {
		ID string `json:"id"`
	}

	var notionResp NotionResponse
	if err := json.Unmarshal(resp, &notionResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return notionResp.ID, nil
}

func (a *Assignment) Update_Notion(col string, value string, obj map[string]string) (err error) {

	assign := a.ToMap()

	var req interface{}

	switch col {

	case "course_code":
		courseReq := types.UpdateCourseCodeRequest{}
		courseReq.Properties = types.PropertiesWithRequiredCourseCode{}
		courseReq.Properties.Courses = types.Courses{
			ID:   "w%3FC%3B",
			Type: "relation",
			Relation: []types.Relation{
				{
					ID: value,
				},
			},
		}
		req = courseReq

	case "deadline":
		deadlineReq := types.UpdateDeadlineRequest{}
		deadlineReq.Properties = types.PropertiesWithRequiredDeadline{}

		dateObj := types.DateObject{
			Start: value,
		}

		deadlineReq.Properties.Deadline = types.Deadline{
			ID:   "_UjC",
			Type: "date",
			Date: &dateObj,
		}

		req = deadlineReq

	case "link":
		linkReq := types.UpdateLinkRequest{}
		linkReq.Properties = types.PropertiesWithRequiredLink{}

		linkReq.Properties.Link = types.Link{
			ID:   "jgPD",
			Type: "url",
			URL:  value,
		}

		req = linkReq

	case "title":
		titleReq := types.UpdateTitleRequest{}
		titleReq.Properties = types.PropertiesWithRequiredName{}

		richTextObj := types.RichText{
			Type: "text",
			Text: &types.TextContent{
				Content: value,
				Link:    nil,
			},
			Annotations: &types.TextAnnotation{
				Bold:          false,
				Italic:        false,
				Strikethrough: false,
				Underline:     false,
				Code:          false,
				Color:         "default",
			},
			PlainText: value,
			Href:      nil,
		}

		titleReq.Properties.AssignmentName = types.AssignmentName{
			Title: []types.RichText{richTextObj},
		}

		req = titleReq

	case "todo":
		todoReq := types.UpdateTODORequest{}
		todoReq.Properties = types.PropertiesWithRequiredTODO{}

		richTextObj := types.RichText{
			Type: "text",
			Text: &types.TextContent{
				Content: value,
				Link:    nil,
			},
			Annotations: &types.TextAnnotation{
				Bold:          false,
				Italic:        false,
				Strikethrough: false,
				Underline:     false,
				Code:          false,
				Color:         "default",
			},
			PlainText: value,
			Href:      nil,
		}

		todoReq.Properties.TODO = types.TODO{
			ID:       "%5DJfC",
			Type:     "rich_text",
			RichText: []types.RichText{richTextObj},
		}

		req = todoReq

	case "type_name":

		typeReq := types.UpdateTypeRequest{}
		typeReq.Properties = types.PropertiesWithRequiredType{}

		typeReq.Properties.Type = types.Type{
			ID:     "S~Ce",
			Type:   "select",
			Select: obj,
		}

		req = typeReq

	case "status_name":

		var statusObj types.StatusObject
		statusObj.ID = obj["id"]
		statusObj.Name = obj["name"]
		statusObj.Color = obj["color"]

		statusReq := types.UpdateStatusRequest{}
		statusReq.Properties = types.PropertiesWithRequiredStatus{}
		statusReq.Properties.Status = types.Status{
			ID:     "%5Bm%5Cs",
			Type:   "status",
			Status: &statusObj,
		}

		req = statusReq

	}

	if req == nil {
		return fmt.Errorf("invalid column type: %s", col)
	}

	url := fmt.Sprintf("pages/%s", assign["notion_id"])

	_, err = notion.SendNotionRequest(req, "PATCH", url, a.User.NotionAPIKey)

	return err
}

func (a *Assignment) Delete_Notion() (err error) {

	assign := a.ToMap()

	req := types.DeletionRequest{}
	req.Archived = true

	_, err = notion.SendNotionRequest(req, "PATCH", assign["notion_id"], a.User.NotionAPIKey)

	return err
}

func GetPage(page_id, sender_id string) (respBody []byte, err error) {

	url := fmt.Sprintf("pages/%s", page_id)
	respBody, err = notion.SendNotionRequest(nil, "GET", url, sender_id)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return respBody, nil
}

func GetPageProperties(page_id, property_id, sender_id string) (respBody []byte, err error) {

	url := fmt.Sprintf("pages/%s/properties/%s", page_id, property_id)
	respBody, err = notion.SendNotionRequest(nil, "GET", url, sender_id)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return respBody, nil
}
