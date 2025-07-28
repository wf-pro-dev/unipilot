package types

type NotionWebhookPayload struct {
	ID      string `json:"id"`
	Authors []struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	} `json:"authors"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Entity    struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	} `json:"entity"`
	Data struct {
		Parent struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		} `json:"parent"`
		Properties []string `json:"updated_properties"`
	} `json:"data"`
}
