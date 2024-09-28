package v0

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Emotion represents a emotion.
type Emotion struct {
	ID              int `gorm:"primaryKey"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UserID          string `gorm:"type:text;not null;index:idx_user_id"`
	Emoji           string `gorm:"type:text;not null"`
	Description     string `gorm:"type:text;not null;default:''"`
	Score           int    `gorm:"type:integer"`
	Task            string `gorm:"type:text"`
	TaskCompletedAt *time.Time
}

// TableName returns the table name.
func (e Emotion) TableName() string {
	return "emotions"
}

// CreateEmotion defines the initial migration, which creates the extension.
var CreateEmotion = gormigrate.Migration{
	ID: "2024-09-28:create-emotion",
	Migrate: func(tx *gorm.DB) error {
		return tx.Migrator().AutoMigrate(&Emotion{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&Emotion{})
	},
}
