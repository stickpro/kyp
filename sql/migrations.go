package sql

import (
	"embed"
)

//go:embed sqlite/migrations/*.sql
var MigrationsPostgres embed.FS

type MigrationParameters struct {
	EmbedFs embed.FS
	Path    string
}

func SqliteMigrationParams() MigrationParameters {
	return MigrationParameters{
		Path:    "sqlite/migrations",
		EmbedFs: MigrationsPostgres,
	}
}
