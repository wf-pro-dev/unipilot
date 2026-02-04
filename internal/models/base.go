package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Base struct {
	ID        string `gorm:"primaryKeytype:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = datatypes.NewUUIDv4().String()
	}
	return nil
}
