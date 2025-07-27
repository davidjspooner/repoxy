package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"

	_ "github.com/davidjspooner/repoxy/pkg/repo/docker"
	_ "github.com/davidjspooner/repoxy/pkg/repo/tf"
)

type GlobalOptions struct {
	cmd.LogOptions
}

func main() {
	root := cmd.NewCommand("", "A Repository Proxy ",
		func(ctx context.Context, options *GlobalOptions, args []string) error {
			err := options.LogOptions.SetupSLOG()
			if err != nil {
				return err
			}
			err = cmd.ShowHelpForMissingSubcommand(ctx)
			return err
		}, &GlobalOptions{LogOptions: cmd.LogOptions{Level: "info"}})

	cmd.Root = root
	versionCommand := cmd.VersionCommand()

	subcommands := cmd.Root.SubCommands()
	subcommands.MustAdd(
		versionCommand,
		serveCommand,
	)

	ctx := context.Background()
	err := cmd.Run(ctx, os.Args[1:])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}
