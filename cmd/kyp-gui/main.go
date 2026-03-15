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
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve db path: %v\n", err)
		os.Exit(1)
	}

	db, err := sqlite.InitLocalStorage(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init storage: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to close storage: %v\n", err)
		}
	}()

	app := gui.NewApp(db)

	if err := wails.Run(&options.App{
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
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
