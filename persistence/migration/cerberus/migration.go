package cerberus

import (
	"github.com/go-gormigrate/gormigrate/v2"

	v0 "github.com/omegaatt36/cerberus/persistence/migration/cerberus/v0"
)

// MigrationList is list of migrations.
var MigrationList = []*gormigrate.Migration{
	&v0.CreateEmotion,
}
