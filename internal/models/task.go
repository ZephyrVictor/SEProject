package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID        uint   `gorm:"primaryKey"`
	ImageID   uint   `gorm:"index;not null"`
	Status    string `gorm:"not null"` // queued, processing, done, failed
	ResultURL string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
