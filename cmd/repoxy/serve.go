package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/davidjspooner/go-http-server/pkg/metric"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
	"github.com/davidjspooner/repoxy/pkg/observability"
	"github.com/davidjspooner/repoxy/pkg/repo"
	"github.com/davidjspooner/repoxy/reactui"

	_ "github.com/davidjspooner/go-fs/pkg/storage/localfile"
	_ "github.com/davidjspooner/go-fs/pkg/storage/s3bucket"
)

type ServeOptions struct {
	Config string `flag:"--config,Path to the configuration file"`
}

func serveHttp(ctx context.Context, config *repo.ConfigFile) error {
	serveMux := mux.NewServeMux(
		observability.HTTPLogger(),
		metric.Middleware(),
		&middleware.Recovery{},
	)
	serveMux.Handle("/metrics", metric.Handler())
	fs, err := repo.NewStorageRoot(ctx, config.Storage)
	if err != nil {
		return fmt.Errorf("failed to connect to storage root: %w", err)
	}
	if err := repo.Initialize(ctx, fs, serveMux); err != nil {
		return fmt.Errorf("failed to initialize repository types: %w", err)
	}
	if uiHandler, err := reactui.Handler(); err != nil {
		return fmt.Errorf("failed to load embedded UI: %w", err)
	} else {
		_ = serveMux.Handle("/ui/{path...}", uiHandler)
		_ = serveMux.Handle("/ui", uiHandler)
		_ = serveMux.Handle("/", http.RedirectHandler("/ui/", http.StatusMovedPermanently))
	}
	repo.RegisterUIRoutes(serveMux)
	for _, r := range config.Repositories {
		_, err := repo.NewRepository(ctx, r)
		if err != nil {
			return fmt.Errorf("failed to create repository instance for %s: %w", r.Name, err)
		}
	}
	err = config.Server.ListenAndServe(ctx, serveMux)
	return err
}

var serveCommand = cmd.NewCommand(
	"serve",
	"Start the repository proxy server",
	func(ctx context.Context, options *ServeOptions, args []string) error {
		config, err := repo.LoadConfigs(options.Config)
		if err != nil {
			return fmt.Errorf("failed to load repository configurations: %w", err)
		}
		err = serveHttp(ctx, config)
		return err
	},
	&ServeOptions{
		Config: "config.yaml",
	},
)
