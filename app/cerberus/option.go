package cerberus

import (
	"github.com/omegaatt36/cerberus/domain"
)

// Option defines jwt option.
type Option interface {
	apply(*Bot)
}

// WithAIServiceOption defines the option to set AIService.
type WithAIServiceOption struct {
	AIService domain.AIService
}

func (o *WithAIServiceOption) apply(bot *Bot) {
	bot.aiService = o.AIService
}

type WithEmotionRepositoryOption struct {
	EmotionRepository domain.EmotionRepository
}

func (o *WithEmotionRepositoryOption) apply(bot *Bot) {
	bot.emotionRepo = o.EmotionRepository
}
