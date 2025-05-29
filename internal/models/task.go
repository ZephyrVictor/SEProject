package models

import (
	"time"

	"gorm.io/gorm"
)

type Task struct {
	ID           string  `gorm:"primaryKey;size:255"`
	ParentID     *string `gorm:"size:255"`
	UserEmail    string  `gorm:"index"`
	Function     string
	Prompt       string
	BaseImageURL string
	ResultURL    *string
	Status       string // QUEUED, PENDING, SUCCEEDED, FAILED
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func MigrateTask(db *gorm.DB) error {
	return db.AutoMigrate(&Task{})
}
