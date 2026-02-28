package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/stickpro/kyp/internal/storage/sqlite/repo_entry"
	"github.com/stickpro/kyp/internal/storage/sqlite/repo_sync"
	"github.com/stickpro/kyp/internal/storage/sqlite/repo_vault"
	kypsql "github.com/stickpro/kyp/sql"
	_ "modernc.org/sqlite"
)

type ILocalStorage interface {
	Entries() repo_entry.Querier
	Vault() repo_vault.Querier
	Sync() repo_sync.Querier
	Close() error
}

type localStorage struct {
	db      *sql.DB
	entries *repo_entry.Queries
	vault   *repo_vault.Queries
	sync    *repo_sync.Queries
}

func InitLocalStorage(dbPath string) (ILocalStorage, error) {
	dsn := fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)",
		dbPath,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &localStorage{
		db:      db,
		entries: repo_entry.New(db),
		vault:   repo_vault.New(db),
		sync:    repo_sync.New(db),
	}, nil
}

func runMigrations(db *sql.DB) error {
	params := kypsql.SqliteMigrationParams()
	goose.SetBaseFS(params.EmbedFs)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, params.Path); err != nil {
		return fmt.Errorf("up: %w", err)
	}

	return nil
}

func (s *localStorage) Entries() repo_entry.Querier { return s.entries }
func (s *localStorage) Vault() repo_vault.Querier   { return s.vault }
func (s *localStorage) Sync() repo_sync.Querier     { return s.sync }

func (s *localStorage) Close() error {
	return s.db.Close()
}
