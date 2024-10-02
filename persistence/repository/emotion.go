package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/omegaatt36/cerberus/domain"
)

var _ domain.EmotionRepository = (*GORMRepository)(nil)

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

// CreateEmotion creates a new emotion.
func (r *GORMRepository) CreateEmotion(ctx context.Context, req domain.CreateEmotionRequest) (int, error) {
	emotion := Emotion{
		UserID:      req.UserID,
		Emoji:       req.Emoji,
		Description: req.Description,
	}

	if err := r.db.Create(&emotion).Error; err != nil {
		return 0, fmt.Errorf("failed to create emotion: %v", err)
	}

	return emotion.ID, nil
}

// UpdateEmotion updates an emotion.
func (r *GORMRepository) UpdateEmotion(ctx context.Context, id int, req domain.UpdateEmotionRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		emotion := Emotion{}
		if err := tx.First(&emotion, id).Error; err != nil {
			return fmt.Errorf("failed to find emotion: %v", err)
		}

		if req.Emoji != nil {
			emotion.Emoji = *req.Emoji
		}
		if req.Description != nil {
			emotion.Description = *req.Description
		}
		if req.Score != nil {
			emotion.Score = *req.Score
		}
		if req.Task != nil {
			emotion.Task = *req.Task
		}
		if req.TaskCompletedAt != nil {
			emotion.TaskCompletedAt = req.TaskCompletedAt
		}
		return tx.Save(&emotion).Error
	})
}
