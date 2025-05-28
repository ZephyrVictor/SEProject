package models

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	ID            uint   `gorm:"primaryKey"`
	UserID        uint   `gorm:"index;not null"`
	ParentID      *uint  `gorm:"index"`
	URL           string `gorm:"not null"`
	Prompt        string
	Operations    string
	AllowDownload bool `gorm:"default:false"`
	AllowRemix    bool `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
