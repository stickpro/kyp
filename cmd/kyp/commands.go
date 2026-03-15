package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stickpro/kyp/internal/config"
	"github.com/stickpro/kyp/internal/storage/sqlite"
	tui "github.com/stickpro/kyp/internal/tui/app"
	"github.com/stickpro/kyp/internal/vault"
	"github.com/stickpro/kyp/pkg/cfg"
	"github.com/stickpro/kyp/pkg/logger"
	"github.com/stickpro/kyp/pkg/storage"
	"github.com/urfave/cli/v3"
)

const defaultConfigPath = "config.yaml"

func commands(currentAppVersion, appName, _ string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "start",
			Description: "Start kyp TUI",
			Flags: []cli.Flag{
				cfgPathsFlag(),
				&cli.StringFlag{
					Name:    "db",
					Usage:   "path to the vault database file",
					Sources: cli.EnvVars("KYP_DB_PATH"),
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				conf, err := loadConfig(c.Args().Slice(), c.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() {
					_ = l.Sync()
				}()

				dbPath, err := resolveDBPath(c.String("db"), conf.Storage.DBPath)
				if err != nil {
					return fmt.Errorf("resolve db path: %w", err)
				}

				storage, err := sqlite.InitLocalStorage(dbPath)
				if err != nil {
					return err
				}
				defer func() {
					if err := storage.Close(); err != nil {
						fmt.Fprintf(os.Stderr, "failed to close storage: %v\n", err)
					}
				}()

				v := vault.Init(storage)
				m := tui.New(v, conf.LockTimeout)

				if _, err := tea.NewProgram(&m, tea.WithAltScreen()).Run(); err != nil {
					return err
				}
				return nil
			},
		},
	}
}

// resolveDBPath returns the path to the database by priority:
// 1. flag --db / env KYP_DB_PATH
// 2. config.yaml storage.db_path
// 3. OS default (~/.local/share/kyp/kyp.db)
func resolveDBPath(flagVal, confVal string) (string, error) {
	path := flagVal
	if path == "" {
		path = confVal
	}
	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return "", fmt.Errorf("create db dir: %w", err)
		}
		return path, nil
	}
	return storage.DefaultDBPath()
}

func cfgPathsFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:    "configs",
		Aliases: []string{"c"},
		Usage:   "paths to configuration files, separated by commas (config.yaml,config.prod.yml,.env)",
		Value:   cli.NewStringSlice(defaultConfigPath).Value(),
	}
}

func loadConfig(args, configPaths []string) (*config.Config, error) {
	conf := new(config.Config)
	if err := cfg.Load(conf,
		cfg.WithLoaderConfig(cfg.Config{
			Args:       args,
			Files:      configPaths,
			MergeFiles: true,
		}),
		cfg.WithOptionalFiles(true),
	); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return conf, nil
}

func defaultLoggerOpts(appName, version string) []logger.Option {
	return []logger.Option{
		logger.WithAppName(appName),
		logger.WithAppVersion(version),
	}
}
