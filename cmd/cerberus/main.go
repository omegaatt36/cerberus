package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	slogzap "github.com/samber/slog-zap/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/omegaatt36/cerberus/app"
	"github.com/omegaatt36/cerberus/app/cerberus"
	"github.com/omegaatt36/cerberus/persistence/database"
	"github.com/omegaatt36/cerberus/persistence/repository"
	"github.com/omegaatt36/cerberus/pkg/gemini"
)

var config struct {
	databaseConnectionOption database.ConnectOption
	logLevel                 string

	slackBotToken string
	slackAppToken string

	geminiAPIKey string
	geminiModel  string
}

var (
	geminiService *gemini.Service
	db            *sql.DB
)

func before(ctx *cli.Context) error {
	if err := initSLog(config.logLevel); err != nil {
		return err
	}

	service, err := gemini.NewService(ctx.Context, config.geminiAPIKey, config.geminiModel)
	if err != nil {
		return err
	}

	geminiService = service

	return database.Initialize(config.databaseConnectionOption)
}

func after(_ *cli.Context) error {
	return errors.Join(db.Close(), geminiService.Close(), database.Finalize())
}

func action(ctx context.Context) {
	bot := cerberus.NewBot(config.slackBotToken, config.slackAppToken,
		&cerberus.WithAIServiceOption{AIService: geminiService},
		&cerberus.WithEmotionRepositoryOption{EmotionRepository: repository.NewGORMRepository(database.GetDB())},
	)

	bot.Run(ctx)
}

func initSLog(logLevel string) error {
	level := zapcore.DebugLevel
	if err := level.Set(logLevel); err != nil {
		level = zapcore.DebugLevel // default level
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), level),
	)

	zapLogger := zap.New(core, zap.AddCaller())

	slog.SetDefault(slog.New(slogzap.Option{Level: slog.LevelDebug, Logger: zapLogger}.NewZapHandler()))

	return nil
}

func main() {
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "slack-bot-token",
			EnvVars:     []string{"SLACK_BOT_TOKEN"},
			Value:       "",
			Required:    true,
			Destination: &config.slackBotToken,
		},
		&cli.StringFlag{
			Name:        "slack-app-token",
			EnvVars:     []string{"SLACK_APP_TOKEN"},
			Value:       "",
			Required:    true,
			Destination: &config.slackAppToken,
		},
		&cli.StringFlag{
			Name:        "gemini-api-key",
			EnvVars:     []string{"GEMINI_API_KEY"},
			Value:       "",
			Required:    true,
			Destination: &config.geminiAPIKey,
		},
		&cli.StringFlag{
			Name:        "gemini-model",
			EnvVars:     []string{"GEMINI_MODEL"},
			Value:       "gemini-1.5-flash",
			Required:    false,
			Destination: &config.geminiModel,
		},
	}
	cliFlags = append(cliFlags, config.databaseConnectionOption.CliFlags()...)

	server := &app.App{
		Action: action,
		Before: before,
		After:  after,
		Flags:  cliFlags,
	}

	server.Run()
}
