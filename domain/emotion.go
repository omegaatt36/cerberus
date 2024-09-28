package domain

import (
	"context"
	"time"
)

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

type CreateEmotionRequest struct {
	UserID      string
	Emoji       string
	Description string
}

type UpdateEmotionRequest struct {
	Emoji           *string
	Description     *string
	Score           *int
	MessagedAt      *time.Time
	Task            *string
	TaskCompletedAt *time.Time
}

type EmotionRepository interface {
	CreateEmotion(ctx context.Context, req CreateEmotionRequest) (int, error)
	UpdateEmotion(ctx context.Context, id int, req UpdateEmotionRequest) error
}
