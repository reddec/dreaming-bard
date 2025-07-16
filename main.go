package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/reddec/dreaming-bard/commands"
)

//nolint:gochecknoglobals
var (
	version = "dev"
)

const (
	description = ``
)

type Config struct {
	ChangeDir string                 `short:"C" help:"Change directory" env:"CHANGE_DIR"`
	Server    commands.ServerCommand `cmd:"" help:"Run server"`
}

func main() {
	var config Config
	ctx := kong.Parse(&config,
		kong.Name("dreaming-bard"),
		kong.Description(fmt.Sprintf("%s %s\n%s\nAuthor: reddec <owner@reddec.net>", "dreaming-bard", version, description)),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	if config.ChangeDir != "" {
		if err := os.Chdir(config.ChangeDir); err != nil {
			slog.Error("change directory", "error", err)
			os.Exit(1)
		}
	}

	err := ctx.Run()
	if err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(2)
	}
}
