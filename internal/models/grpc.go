package models

type Heartbeat struct {
	Body string `json:"body"`
}

type MessageType string

const (
	MessageNoContent  MessageType = "NoContent"
	MessageCourse     MessageType = "Course"
	MessageAssignment MessageType = "Assignment"
	MessageDocument   MessageType = "Document"
	MessageNote       MessageType = "Note"
	MessageInvitation MessageType = "Invitation"
)

type Message struct {
	SenderID   uint        `json:"sender_id"`
	ReceiverID uint        `json:"receiver_id"`
	Title      string      `json:"title"`
	Message    string      `json:"message"`
	Data       []byte      `json:"data"`
	Type       MessageType `json:"type"`
}
