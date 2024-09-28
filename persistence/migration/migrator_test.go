package migration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/omegaatt36/cerberus/persistence/database"
	"github.com/omegaatt36/cerberus/persistence/migration"
	apimigration "github.com/omegaatt36/cerberus/persistence/migration/cerberus"
)

func TestMigrateAPI(t *testing.T) {
	s := assert.New(t)

	finalize := database.TestingInitialize(database.PostgresOpt)
	defer finalize()

	db := database.GetDB()

	mg := migration.NewMigrator(db, []any{}, apimigration.MigrationList)

	s.NoError(mg.Upgrade())
}
