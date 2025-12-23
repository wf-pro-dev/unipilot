package aimessage

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"unipilot/internal/errors"
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
func (m *LocalAiMessage) ToUIMessage() (map[string]interface{}, error) {
	var parts []interface{}
	var metadata map[string]interface{}

	if err := json.Unmarshal(m.Parts, &parts); err != nil {
		return nil, errors.Wrap(err, errors.ProcJSONUnmarshalFailed, "Failed to unmarshal message parts")
	}
	json.Unmarshal(m.Metadata, &metadata)

	return map[string]interface{}{
		"id":        m.ID,
		"role":      string(m.Role),
		"parts":     parts,
		"metadata":  metadata,
		"createdAt": m.CreatedAt,
	}, nil
}

// FromUIMessage creates from AI SDK message
func FromUIMessage(assignmentID uint, uiMessage map[string]interface{}) (*LocalAiMessage, error) {
	message := &LocalAiMessage{
		AssignmentID: assignmentID,
	}

	if id, ok := uiMessage["id"].(string); ok {
		message.ID = id
	} else {
		return nil, errors.NewAppError(errors.ValidationInvalid, "Invalid message ID", nil)
	}

	if role, ok := uiMessage["role"].(string); ok {
		message.Role = Roles(role)
	} else {
		return nil, errors.NewAppError(errors.ValidationInvalid, "Invalid message role", nil)
	}

	// Store parts as JSON
	if parts, ok := uiMessage["parts"]; ok {
		partsJSON, err := json.Marshal(parts)
		if err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal message parts")
		}
		message.Parts = datatypes.JSON(partsJSON)
	}

	// Store metadata as JSON
	if metadata, ok := uiMessage["metadata"]; ok {
		metadataJSON, err := json.Marshal(metadata)
		if err != nil {
			return nil, errors.Wrap(err, errors.ProcJSONMarshalFailed, "Failed to marshal message metadata")
		}
		message.Metadata = datatypes.JSON(metadataJSON)
	}

	return message, nil
}
