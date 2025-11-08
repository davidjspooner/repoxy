package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-server/pkg/handler"
	"github.com/davidjspooner/go-http-server/pkg/metric"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
	"github.com/davidjspooner/repoxy/pkg/repo"

	_ "github.com/davidjspooner/go-fs/pkg/storage/localfile"
)

type ServeOptions struct {
	Config string `flag:"--config,Path to the configuration file"`
}

var serveCommand = cmd.NewCommand(
	"serve",
	"Start the repository proxy server",
	func(ctx context.Context, options *ServeOptions, args []string) error {
		repoConfig, err := repo.LoadConfigs(options.Config)
		if err != nil {
			return fmt.Errorf("failed to load repository configurations: %w", err)
		}
		serveMux := mux.NewServeMux(loggerMiddleware(), metric.Middleware(), &middleware.Recovery{})
		serveMux.Handle("/metrics", metric.Handler())
		repo.AddAllToMux(serveMux)
		forest, err := repo.NewStorageForest(ctx, repoConfig.Storage)
		if err != nil {
			return fmt.Errorf("failed to connect to storage forest: %w", err)
		}
		for _, r := range repoConfig.Repositories {
			_, err := repo.NewRepository(ctx, forest, r)
			if err != nil {
				return fmt.Errorf("failed to create repository instance for %s: %w", r.Name, err)
			}
		}
		err = repoConfig.Server.ListenAndServe(ctx, serveMux)
		return err
	},
	&ServeOptions{
		Config: "config.yaml",
	},
)

func loggerMiddleware() *middleware.Log {
	return &middleware.Log{
		AfterRequest: func(ctx context.Context, r *http.Request, observed *handler.Observation) {

			attrs := []any{
				slog.String("req_id", observed.Request.ID),
				slog.String("cli_id", r.RemoteAddr),
			}
			if observed.Request.User != "" {
				attrs = append(attrs, slog.String("user", observed.Request.User))
			}
			attrs = append(attrs,
				slog.String("method", r.Method),
				slog.String("url", r.URL.String()),
			)
			if observed.Request.Body.Length > 0 {
				attrs = append(attrs, slog.Int("req_len", observed.Request.Body.Length))
			}
			attrs = append(attrs,
				slog.Int("status", observed.Response.Status),
				slog.Float64("duration", observed.Response.Duration.Seconds()),
			)
			if observed.Response.Body.Length > 0 {
				attrs = append(attrs, slog.Int("resp_len", observed.Response.Body.Length))
			}

			if observed.RoutePattern != "" {
				attrs = append(attrs, slog.String("route", observed.RoutePattern))
			}

			statusType := observed.Response.Status / 100
			switch statusType {
			case 1, 2: //1xx, 2xx
				slog.InfoContext(ctx, "request completed", attrs...)
			case 3: //3xx
				slog.InfoContext(ctx, "request redirected", attrs...)
			case 4: //4xx
				slog.WarnContext(ctx, "client error", attrs...)
			case 5: //5xx
				slog.ErrorContext(ctx, "server error", attrs...)
			default: //other
				slog.ErrorContext(ctx, "unexpected status", attrs...)
			}
		},
	}
}
