package course

import (
	"encoding/json"
	"fmt"
	"os"

	"unipilot/internal/services/notion"
	"unipilot/internal/types"
)

var NOTION_API_KEY = os.Getenv("NOTION_API_KEY")

const BASE_URL = "https://api.notion.com/v1"

// AddAssignmentToNotion adds an assignment to Notion efficiently
func (c *Course) Add_Notion() (string, error) {

	// Create a single rich text object for reuse
	createRichTextObj := func() types.RichText {
		text := types.RichText{
			Type: "text",
			Text: &types.TextContent{
				Content: "",
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
			PlainText: "",
			Href:      nil,
		}
		return text
	}

	// Create the request with strongly typed fields
	req := types.CoursePageRequest{}
	req.Parent.Type = "database_id"
	req.Parent.DatabaseID = c.User.CoursesDbId

	properties := types.CourseProperties{
		Name: types.Title{
			Type: "title",
		},
		Code: types.Text{
			Type: "rich_text",
		},
		RoomNumber: types.Text{
			Type: "rich_text",
		},
		Duration: types.Text{
			Type: "rich_text",
		},
	}

	req.Properties = &properties

	// Set name
	name_obj := types.Title{
		Type: "title",
	}
	nameText := createRichTextObj()
	nameText.Text.Content = c.Name
	nameText.PlainText = c.Name
	name_obj.Title = []types.RichText{nameText}
	req.Properties.Name = name_obj

	// Set code
	code_obj := types.Text{
		Type: "rich_text",
	}
	codeText := createRichTextObj()
	codeText.Text.Content = c.Code
	codeText.PlainText = c.Code
	code_obj.RichText = []types.RichText{codeText}
	req.Properties.Code = code_obj

	// Set name
	room_number_obj := types.Text{
		Type: "rich_text",
	}
	roomNumberText := createRichTextObj()
	roomNumberText.Text.Content = c.RoomNumber
	roomNumberText.PlainText = c.RoomNumber
	room_number_obj.RichText = []types.RichText{roomNumberText}
	req.Properties.RoomNumber = room_number_obj

	// Set duration
	duration_obj := types.Text{
		Type: "rich_text",
	}
	durationText := createRichTextObj()
	durationText.Text.Content = c.Duration
	durationText.PlainText = c.Duration
	duration_obj.RichText = []types.RichText{durationText}
	req.Properties.Duration = duration_obj

	resp, err := notion.SendNotionRequest(req, "POST", "pages", c.User.NotionAPIKey)
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
