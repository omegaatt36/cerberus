package domain

import "context"

// AIService defines the interface for AI interactions
type AIService interface {
	GetEmotionScore(ctx context.Context, input string) (int, error)
	GenerateTaskSuggestion(ctx context.Context, emoji string, description string, score int) (string, error)
	GenerateDailySummary(ctx context.Context, averageScore float64) (string, error)
}
