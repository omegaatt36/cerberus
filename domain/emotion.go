package domain

import (
	"context"
	"time"
)

// Emotion represents an emotional state with associated metadata
type Emotion struct {
	ID              int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UserID          string
	Emoji           string
	Description     string
	Score           int
	MessagedAt      *time.Time
	Task            string
	TaskCompletedAt *time.Time
}

// CreateEmotionRequest represents the data required to create a new Emotion
type CreateEmotionRequest struct {
	UserID      string
	Emoji       string
	Description string
}

// UpdateEmotionRequest represents the data that can be updated for an existing Emotion
type UpdateEmotionRequest struct {
	Emoji           *string
	Description     *string
	Score           *int
	MessagedAt      *time.Time
	Task            *string
	TaskCompletedAt *time.Time
}

// EmotionRepository defines the interface for Emotion data persistence
type EmotionRepository interface {
	CreateEmotion(ctx context.Context, req CreateEmotionRequest) (int, error)
	UpdateEmotion(ctx context.Context, id int, req UpdateEmotionRequest) error
}
