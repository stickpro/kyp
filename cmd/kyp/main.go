package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/urfave/cli/v3"
)

var (
	appName    = "kyp"
	version    = "local"
	commitHash = "unknown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), []os.Signal{
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL,
	}...)
	defer cancel()

	app := &cli.Command{
		Name:        appName,
		Description: "Keep Your Passwords - password manager",
		Version:     getBuildVersion(),
		Suggest:     true,
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
		},
		Commands:       commands(version, appName, commitHash),
		DefaultCommand: "start",
	}

	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Println(err.Error())
	}
}

func getBuildVersion() string {
	return fmt.Sprintf(
		"\n\nrelease: %s\ncommit hash: %s\ngo version: %s",
		version,
		commitHash,
		runtime.Version(),
	)
}
