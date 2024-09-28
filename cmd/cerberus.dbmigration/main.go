package main

import (
	"context"
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/omegaatt36/cerberus/app"
	"github.com/omegaatt36/cerberus/persistence/database"
	"github.com/omegaatt36/cerberus/persistence/migration"
	"github.com/omegaatt36/cerberus/persistence/migration/cerberus"
)

var config struct {
	databaseConnectionOption database.ConnectOption
	rollback                 bool
}

func before(_ *cli.Context) error {
	return database.Initialize(config.databaseConnectionOption)
}

func after(_ *cli.Context) error {
	return database.Finalize()
}

func action(ctx context.Context) {
	db := database.GetDB().Debug()
	mg := migration.NewMigrator(db, []any{}, cerberus.MigrationList)

	if config.rollback {
		if err := mg.Rollback(); err != nil {
			slog.Error("rollback error", slog.String("error", err.Error()))
			panic(err)
		}

		return
	}

	if err := mg.Upgrade(); err != nil {
		slog.Error("upgrade error", slog.String("error", err.Error()))
		panic(err)
	}
}

func main() {
	cliFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "rollback-last",
			EnvVars:     []string{"ROLLBACK_LAST"},
			Value:       false,
			Destination: &config.rollback,
		}}
	cliFlags = append(cliFlags, config.databaseConnectionOption.CliFlags()...)

	server := app.App{
		Action: action,
		Before: before,
		After:  after,
		Flags:  cliFlags,
	}

	server.Run()
}
