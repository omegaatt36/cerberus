package cerberus

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	"github.com/omegaatt36/cerberus/domain"
)

// Bot represents the Slack bot with its configuration and dependencies
type Bot struct {
	slackBotToken string
	slackAppToken string

	slackClient  *slack.Client
	socketClient *socketmode.Client

	emotionRepo domain.EmotionRepository
	aiService   domain.AIService
}

// NewBot creates a new Bot instance.
func NewBot(slackBotToken, slackAppToken string, options ...Option) *Bot {
	bot := &Bot{
		slackBotToken: slackBotToken,
		slackAppToken: slackAppToken,
	}

	for _, option := range options {
		option.apply(bot)
	}

	bot.slackClient = slack.New(bot.slackBotToken, slack.OptionAppLevelToken(bot.slackAppToken))
	bot.socketClient = socketmode.New(bot.slackClient)

	return bot
}

// Run starts the bot and listens for Slack events
func (b *Bot) Run(ctx context.Context) {
	go b.handleEvents(ctx)

	slog.Info("Starting to listen for Slack events")
	if err := b.socketClient.RunContext(ctx); err != nil {
		slog.Error("Error while listening", "error", err)
	} else {
		slog.Info("Bot stopped listening without error")
	}

	slog.Info("Bot execution completed")
}

func parseInput(input string) (emoji string, description string) {
	parts := strings.SplitN(input, " ", 2)

	if len(parts) == 0 {
		return "", ""
	}

	emoji = parts[0]
	if !isValidSlackEmoji(emoji) {
		return "", ""
	}

	if len(parts) == 1 {
		return emoji, ""
	}

	return emoji, parts[1]
}

func isValidSlackEmoji(s string) bool {
	if len(s) < 3 || s[0] != ':' || s[len(s)-1] != ':' {
		return false
	}

	for _, c := range s[1 : len(s)-1] {
		if !isValidEmojiChar(c) {
			return false
		}
	}

	return true
}

func isValidEmojiChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

func (b *Bot) handleEmojiCommand(ctx context.Context, command *slack.SlashCommand) (string, error) {
	slog.Info("Starting handleEmojiCommand")
	slog.Info("Command", "command", command)

	if command == nil {
		return "", fmt.Errorf("error: received nil command")
	}

	input := command.Text
	if input == "" {
		slog.Info("Empty input received")
		return "Please provide an emoji and optional text.", nil
	}

	emoji, description := parseInput(input)
	if emoji == "" {
		return "", fmt.Errorf("please provide a valid emoji at the beginning of your message.\n (e.g., /emoji 😊 Feeling optimistic today!)")
	}

	userID := command.UserID
	if userID == "" {
		return "", fmt.Errorf("can't find user ID")
	}

	id, err := b.emotionRepo.CreateEmotion(ctx, domain.CreateEmotionRequest{
		UserID:      userID,
		Emoji:       emoji,
		Description: description,
	})
	if err != nil {
		slog.ErrorContext(ctx, "error storing initial data", "error", err)
		return "", fmt.Errorf("error processing your request, please try again")
	}

	score, err := b.aiService.GetEmotionScore(context.Background(), input)
	if err != nil {
		return "", fmt.Errorf("analyzing emotion score failed: %w", err)
	}

	if err := b.emotionRepo.UpdateEmotion(ctx, id, domain.UpdateEmotionRequest{Score: &score}); err != nil {
		slog.ErrorContext(ctx, "error updating score", "error", err)
	}

	task, err := b.aiService.GenerateTaskSuggestion(ctx, emoji, description, score)
	if err != nil {
		return "", fmt.Errorf("generating task suggestion failed: %w", err)
	}

	if err := b.emotionRepo.UpdateEmotion(ctx, id, domain.UpdateEmotionRequest{Task: &task}); err != nil {
		slog.ErrorContext(ctx, "updating task failed", "error", err)
	}

	return task, nil
}

func (b *Bot) handleEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping event handling")
			return
		case event := <-b.socketClient.Events:
			switch event.Type {
			case socketmode.EventTypeConnecting:
				slog.Info("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				slog.Info("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				slog.Info("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					slog.Info("ignored event", "event", event)
					continue
				}
				slog.Info("event received", "event", eventsAPIEvent)
				b.socketClient.Ack(*event.Request)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := event.Data.(slack.SlashCommand)
				if !ok {
					slog.Info("ignored event", "event", event)
					continue
				}
				b.socketClient.Ack(*event.Request)
				if err := b.handleSlashCommand(cmd); err != nil {
					slog.Error("Error handling slash command", "error", err)
				}
			case socketmode.EventTypeHello:
				slog.Info("Received hello event from Slack")
			default:
				slog.Info("Unexpected event type received", "type", event.Type)
			}
		}
	}
}

func (b *Bot) handleSlashCommand(command slack.SlashCommand) error {
	ctx := context.Background()

	slog.With(
		"command", command.Command,
		"text", command.Text,
		"user_id", command.UserID,
		"channel_id", command.ChannelID,
	).InfoContext(ctx, "Handling slash command")

	switch command.Command {
	case "/emoji":
		channelID := command.ChannelID
		if _, _, err := b.socketClient.PostMessageContext(ctx, channelID,
			slack.MsgOptionText(fmt.Sprintf("<@%s> said: %s", command.UserID, command.Text), false)); err != nil {
			slog.ErrorContext(ctx, "error sending message", "error", err)
		}
		message, err := b.handleEmojiCommand(ctx, &command)
		if err != nil {
			return b.sendMessage(ctx, channelID, err.Error())
		}
		return b.sendMessage(ctx, channelID, message)
	default:
		return fmt.Errorf("unknown command: %s", command.Command)
	}
}

func (b *Bot) sendMessage(ctx context.Context, channelID string, message string) error {
	_, _, err := b.socketClient.PostMessageContext(ctx, channelID, slack.MsgOptionText(message, false))
	return err
}
