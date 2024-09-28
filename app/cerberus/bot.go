package cerberus

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	_ "github.com/mattn/go-sqlite3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"google.golang.org/api/option"

	"github.com/omegaatt36/cerberus/domain"
)

type Bot struct {
	slackBotToken string
	slackAppToken string

	slackClient  *slack.Client
	socketClient *socketmode.Client

	emotionRepo domain.EmotionRepository
	aiService   domain.AIService
}

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

func (b *Bot) Run(ctx context.Context) {
	slog.Info("Bot configuration complete")
	slog.Info("Bot Token", "token", b.slackBotToken[:10]+"...")
	slog.Info("App Token", "token", b.slackAppToken[:10]+"...")
	slog.Info("Database initialized successfully")
	slog.Info("Preparing to start bot...")

	go b.handleEvents()

	slog.Info("Starting to listen for Slack events")
	if err := b.socketClient.RunContext(ctx); err != nil {
		slog.Error("Error while listening", "error", err)
	} else {
		slog.Info("Bot stopped listening without error")
	}

	slog.Info("Bot execution completed")
}

func parseInput(input string) (string, string) {
	// Split the input string by space
	parts := strings.SplitN(input, " ", 2)

	// If there's only one part, it's just the emoji
	if len(parts) == 1 {
		return parts[0], ""
	}

	// Otherwise, return emoji and description
	return parts[0], parts[1]
}

func (bot *Bot) handleEmojiCommand(command *slack.SlashCommand) {
	ctx := context.Background()
	slog.Info("Starting handleEmojiCommand")
	slog.Info("Command", "command", command)

	if command == nil {
		slog.Error("Error: Received nil command")
		return
	}

	input := command.Text
	slog.Info("Input received", "input", input)
	if input == "" {
		slog.Info("Empty input received")
		_, _, err := bot.slackClient.PostMessage(command.ChannelID, slack.MsgOptionText("Please provide an emoji and optional text.", false))
		if err != nil {
			slog.Error("Error sending message", "error", err)
			return
		}
		return
	}

	emoji, description := parseInput(input)
	if emoji == "" {
		slog.Error("Error: No emoji found in input")
		_, _, err := bot.slackClient.PostMessage(command.ChannelID, slack.MsgOptionText("Please provide a valid emoji.", false))
		if err != nil {
			slog.Error("Error sending message", "error", err)
		}
		return
	}

	userID := command.UserID
	if userID == "" {
		slog.Error("Error: Empty user ID")
		return
	}

	// Stage 1: Store initial data
	id, err := bot.emotionRepo.CreateEmotion(ctx, domain.CreateEmotionRequest{
		UserID:      userID,
		Emoji:       emoji,
		Description: description,
	})
	if err != nil {
		slog.Error("Error storing initial data", "error", err)
		_, _, err := bot.slackClient.PostMessage(command.ChannelID, slack.MsgOptionText("Error processing your request. Please try again.", false))
		if err != nil {
			slog.Error("Error sending message", "error", err)
		}
		return
	}

	// Stage 2: Get and update emotion score
	slog.Info("Analyzing emotion score")
	score, err := bot.aiService.GetEmotionScore(context.Background(), input)
	if err != nil {
		slog.Error("Error analyzing emotion", "error", err)
	} else {
		if err := bot.emotionRepo.UpdateEmotion(ctx, id, domain.UpdateEmotionRequest{Score: &score}); err != nil {
			slog.Error("Error updating score", "error", err)
		}
	}

	// Stage 3: Generate and update task suggestion
	slog.Info("Generating task suggestion")
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		slog.Error("Error creating Gemini client", "error", err)
	} else {
		defer client.Close()
		task, err := bot.aiService.GenerateTaskSuggestion(ctx, emoji, description, score)
		if err != nil {
			slog.Error("Error generating task suggestion", "error", err)
		} else {

			if err := bot.emotionRepo.UpdateEmotion(ctx, id, domain.UpdateEmotionRequest{Task: &task}); err != nil {
				slog.Error("Error updating task", "error", err)
			}
			message := fmt.Sprintf("Emotion score for '%s %s': %d\n建議任務: %s", emoji, description, score, task)
			_, _, err = bot.slackClient.PostMessage(command.ChannelID, slack.MsgOptionText(message, false))
			if err != nil {
				slog.Error("Error sending message", "error", err)
			}
		}
	}
}

func (bot *Bot) handleEvents() {
	for evt := range bot.socketClient.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			slog.Info("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			slog.Info("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			slog.Info("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
			if !ok {
				slog.Info("Ignored event", "event", evt)
				continue
			}
			slog.Info("Event received", "event", eventsAPIEvent)
			bot.socketClient.Ack(*evt.Request)
		case socketmode.EventTypeSlashCommand:
			cmd, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				slog.Info("Ignored event", "event", evt)
				continue
			}
			bot.socketClient.Ack(*evt.Request)
			if err := bot.handleSlashCommand(cmd); err != nil {
				slog.Error("Error handling slash command", "error", err)
			}
		default:
			slog.Info("Unexpected event type received", "type", evt.Type)
		}
	}
}
func (bot *Bot) handleSlashCommand(command slack.SlashCommand) error {
	slog.Info("Handling slash command", "command", command.Command)
	slog.Info("Command text", "text", command.Text)
	slog.Info("User ID", "user_id", command.UserID)
	slog.Info("Channel ID", "channel_id", command.ChannelID)

	slog.Info("Received command", "command", command)

	switch command.Command {
	case "/emoji":
		bot.handleEmojiCommand(&command)
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command.Command)
	}
}
