package gemini

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Service is a wrapper around the Gemini client
type Service struct {
	model  string
	client *genai.Client
}

// NewService creates a new Gemini service
func NewService(apiKey, model string) (*Service, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &Service{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client
func (g *Service) Close() error {
	return g.client.Close()
}

// GetEmotionScore analyzes the emotion of a given input string or emoji
func (g *Service) GetEmotionScore(ctx context.Context, input string) (int, error) {
	const formatGetEmotionScorePrompt = `Analyze the emotion in the following text or emoji and provide a score from 0 to 100, where 0 is very negative and 100 is very positive. Only respond with the number, no other text. Text to analyze: %s`

	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(fmt.Sprintf(formatGetEmotionScorePrompt, input)))
	if err != nil {
		return 0, fmt.Errorf("error generating content: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return 0, fmt.Errorf("no response received from Gemini")
	}

	scoreStr := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			scoreStr += string(textPart)
		}
	}

	score, err := strconv.Atoi(strings.TrimSpace(scoreStr))
	if err != nil {
		return 0, fmt.Errorf("failed to parse score: %v", err)
	}

	if score < 0 || score > 100 {
		return 0, fmt.Errorf("invalid score received: %d", score)
	}

	return score, nil
}

// GenerateTaskSuggestion generates a task suggestion based on the emotion score and description
func (g *Service) GenerateTaskSuggestion(ctx context.Context, emoji string, description string, score int) (string, error) {
	const formatGenerateTaskSuggestionPrompt = `Based on the emoji %s, description '%s', and emotion score %d (0-100, where 0 is very negative and 100 is very positive), suggest a task in Traditional Chinese that can improve mood in an office setting. Provide only one short suggestion, no numbering or explanation.`
	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(fmt.Sprintf(formatGenerateTaskSuggestionPrompt, emoji, description, score)))
	if err != nil {
		return "", fmt.Errorf("failed to generate task suggestion: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response received for task suggestion")
	}

	var suggestion string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			suggestion += string(textPart)
		}
	}

	return strings.TrimSpace(suggestion), nil
}

// GenerateDailySummary generates a summary using Gemini based on the average score
func (g *Service) GenerateDailySummary(ctx context.Context, average float64) (string, error) {
	const formatGenerateDailySummaryPrompt = `Based on the average emotion score of %.2f (0-100, where 0 is very negative and 100 is very positive), provide a brief summary in Traditional Chinese about the overall mood and a general suggestion for improvement. Keep it concise and positive.`

	resp, err := g.client.GenerativeModel(g.model).GenerateContent(ctx, genai.Text(fmt.Sprintf(formatGenerateDailySummaryPrompt, average)))
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response received for summary")
	}

	var summary string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			summary += string(textPart)
		}
	}

	return strings.TrimSpace(summary), nil
}
