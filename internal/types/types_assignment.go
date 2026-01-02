package types

// TypeInfo represents the type of assignment
type TypeInfo struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// CourseInfo represents the course relation
type CourseID struct {
	ID string `json:"id,omitempty"`
}

// Use struct to define the request structure for better type safety and readability
type TextContent struct {
	Content string      `json:"content"`
	Link    interface{} `json:"link,omitempty"`
}

type TextAnnotation struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

type RichText struct {
	Type        string          `json:"type,omitempty"`
	Text        *TextContent    `json:"text"`
	Annotations *TextAnnotation `json:"annotations"`
	PlainText   string          `json:"plain_text,omitempty"`
	Href        interface{}     `json:"href"`
}

type DateObject struct {
	Start    string      `json:"start,omitempty"`
	End      interface{} `json:"end,omitempty"`
	TimeZone interface{} `json:"time_zone,omitempty"`
}

type StatusObject struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}
type AssignmentName struct {
	ID    string     `json:"id,omitempty"`
	Type  string     `json:"type,omitempty"`
	Title []RichText `json:"title"`
}

type Deadline struct {
	ID   string      `json:"id,omitempty"`
	Type string      `json:"type,omitempty"`
	Date *DateObject `json:"date,omitempty"`
}

type Relation struct {
	ID string `json:"id,omitempty"`
}

type Courses struct {
	ID       string     `json:"id,omitempty"`
	Type     string     `json:"type,omitempty"`
	Relation []Relation `json:"relation,omitempty"`
	HasMore  bool       `json:"has_more,omitempty"`
}

type Type struct {
	ID     string            `json:"id,omitempty"`
	Type   string            `json:"type,omitempty"`
	Select map[string]string `json:"select,omitempty"`
}

type Status struct {
	ID     string        `json:"id,omitempty"`
	Type   string        `json:"type,omitempty"`
	Status *StatusObject `json:"status,omitempty"`
}

type TODO struct {
	ID       string     `json:"id,omitempty"`
	Type     string     `json:"type,omitempty"`
	RichText []RichText `json:"rich_text,omitempty"`
}

type Link struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Properties struct {
	TODO           TODO           `json:"TODO"`
	AssignmentName AssignmentName `json:"Assignment Name"`
	Deadline       Deadline       `json:"Deadline"`
	Courses        Courses        `json:"Courses"`
	Type           Type           `json:"Type"`
	Status         Status         `json:"Status"`
	Link           Link           `json:"Link"`
}

type PageRequest struct {
	Cover  *interface{} `json:"cover,omitempty"`
	Icon   *interface{} `json:"icon,omitempty"`
	Parent struct {
		Type       string `json:"type,omitempty"`
		DatabaseID string `json:"database_id,omitempty"`
	} `json:"parent,omitempty"`
	Archived   bool        `json:"archived,omitempty"`
	InTrash    bool        `json:"in_trash,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
}

// PropertiesWithRequiredName is a copy of Properties where AssignmentName is always included
type PropertiesWithRequiredName struct {
	AssignmentName AssignmentName `json:"Assignment Name"` // No omitempty here
}

type UpdateTitleRequest struct {
	Properties PropertiesWithRequiredName `json:"properties,omitempty"`
}

type PropertiesWithRequiredTODO struct {
	TODO TODO `json:"TODO"` // No omitempty here
}

type UpdateTODORequest struct {
	Properties PropertiesWithRequiredTODO `json:"properties,omitempty"`
}

type PropertiesWithRequiredDeadline struct {
	Deadline Deadline `json:"Deadline"` // No omitempty here
}

type UpdateTypeRequest struct {
	Properties PropertiesWithRequiredType `json:"properties,omitempty"`
}

type PropertiesWithRequiredType struct {
	Type Type `json:"Type"` // No omitempty here
}

type UpdateDeadlineRequest struct {
	Properties PropertiesWithRequiredDeadline `json:"properties,omitempty"`
}

type PropertiesWithRequiredLink struct {
	Link Link `json:"Link"` // No omitempty here
}

type UpdateLinkRequest struct {
	Properties PropertiesWithRequiredLink `json:"properties,omitempty"`
}

type PropertiesWithRequiredCourseCode struct {
	Courses Courses `json:"Courses"` // No omitempty here
}

type UpdateCourseCodeRequest struct {
	Properties PropertiesWithRequiredCourseCode `json:"properties,omitempty"`
}

type PropertiesWithRequiredStatus struct {
	Status Status `json:"Status"` // No omitempty here
}

type UpdateStatusRequest struct {
	Properties PropertiesWithRequiredStatus `json:"properties,omitempty"`
}

type DeletionRequest struct {
	Archived bool `json:"archived,omitempty"`
}

// COLUMNS is a list of structs that contain the ID and Name of the column
var COLUMNS = map[string]string{
	"id":        "id",
	"w%3FC%3B":  "course_code",
	"S~Ce":      "type",
	"title":     "title",
	"_UjC":      "deadline",
	"%5DJfC":    "todo",
	"notion_id": "notion_id",
	"jgPD":      "link",
	"%5Bm%5Cs":  "status",
}

var DEFAULT_COLUMNS_FOR_LS = []string{"id", "type", "course_code", "title", "deadline", "todo", "status"}

func GetColumnFromPropertyId(property_id string) (column string, err error) {

	return column, nil
}
