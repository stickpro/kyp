package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/stickpro/kyp/internal/gui"
	"github.com/stickpro/kyp/internal/storage/sqlite"
	"github.com/stickpro/kyp/pkg/storage"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	db, err := sqlite.InitLocalStorage(dbPath)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close storage: %v\n", err)
		}
	}()

	app := gui.NewApp(db)

	return wails.Run(&options.App{
		Title:  "kyp",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
	})
}
