package models

import (
	"time"

	"gorm.io/gorm"
)

type Favorite struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	ImageID   uint `gorm:"index;not null"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
