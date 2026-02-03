package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Base struct {
	ID        datatypes.UUID `gorm:"primaryKeytype:uuid" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
