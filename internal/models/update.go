package models

import (
	"time"

	"gorm.io/gorm"
)

type LocalSync struct {
	gorm.Model
	Entity      Entity    `gorm:"not null"`
	EntityID    uint      `gorm:"not null"`
	Action      string    `gorm:"not null"`
	Column      string    `gorm:"not null"`
	Value       string    `gorm:"not null"`
	RetryCount  int       `gorm:"default:0"`
	NextRetryAt time.Time `gorm:"not null"`
	LastError   string    `gorm:"not null"`
}
