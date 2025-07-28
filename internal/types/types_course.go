package types

// Use struct to define the request structure for better type safety and readability

type Title struct {
	ID    string     `json:"id,omitempty"`
	Type  string     `json:"type,omitempty"`
	Title []RichText `json:"title"`
}

type Text struct {
	ID       string     `json:"id,omitempty"`
	Type     string     `json:"type,omitempty"`
	RichText []RichText `json:"rich_text"`
}

type CourseProperties struct {
	Name       Title `json:"Course Name,omitempty"`
	Code       Text  `json:"Course Code,omitempty"`
	RoomNumber Text  `json:"Room Number,omitempty"`
	Duration   Text  `json:"Course Duration,omitempty"`
}

type CoursePageRequest struct {
	Cover  *interface{} `json:"cover,omitempty"`
	Icon   *interface{} `json:"icon,omitempty"`
	Parent struct {
		Type       string `json:"type,omitempty"`
		DatabaseID string `json:"database_id,omitempty"`
	} `json:"parent,omitempty"`
	Archived   bool        `json:"archived,omitempty"`
	InTrash    bool        `json:"in_trash,omitempty"`
	Properties *CourseProperties `json:"properties,omitempty"`
}
