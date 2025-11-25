package aimessage

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Roles string

const (
	RoleSystem    Roles = "system"
	RoleUser      Roles = "user"
	RoleAssistant Roles = "assistant"
)

// LocalAiMessage maps to the UIMessage interface
type LocalAiMessage struct {
	ID           string         `gorm:"primaryKey"`
	AssignmentID uint           `gorm:"not null;index"`
	Role         Roles          `gorm:"not null;type:varchar(20)"`
	Parts        datatypes.JSON `gorm:"type:jsonb"` // Store entire parts array as JSON
	Metadata     datatypes.JSON `gorm:"type:jsonb"` // Store optional metadata as JSON

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ToUIMessage converts to a format compatible with AI SDK
func (m *LocalAiMessage) ToUIMessage() map[string]interface{} {
	var parts []interface{}
	var metadata map[string]interface{}

	json.Unmarshal(m.Parts, &parts)
	json.Unmarshal(m.Metadata, &metadata)

	return map[string]interface{}{
		"id":        m.ID,
		"role":      string(m.Role),
		"parts":     parts,
		"metadata":  metadata,
		"createdAt": m.CreatedAt,
	}
}

// FromUIMessage creates from AI SDK message
func FromUIMessage(assignmentID uint, uiMessage map[string]interface{}) (*LocalAiMessage, error) {
	message := &LocalAiMessage{
		AssignmentID: assignmentID,
	}

	if id, ok := uiMessage["id"].(string); ok {
		message.ID = id
	}

	if role, ok := uiMessage["role"].(string); ok {
		message.Role = Roles(role)
	}

	// Store parts as JSON
	if parts, ok := uiMessage["parts"]; ok {
		partsJSON, err := json.Marshal(parts)
		if err == nil {
			message.Parts = datatypes.JSON(partsJSON)
		}
	}

	// Store metadata as JSON
	if metadata, ok := uiMessage["metadata"]; ok {
		metadataJSON, err := json.Marshal(metadata)
		if err == nil {
			message.Metadata = datatypes.JSON(metadataJSON)
		}
	}

	return message, nil
}
