package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/davidjspooner/go-http-server/pkg/handler"
	"github.com/davidjspooner/go-http-server/pkg/metric"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
	"github.com/davidjspooner/go-http-server/pkg/mux"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type ServeOptions struct {
}

var serveCommand = cmd.NewCommand(
	"serve",
	"Start the repository proxy server",
	func(ctx context.Context, options *ServeOptions, args []string) error {

		serveMux := mux.NewServeMux(loggerMiddleware(), metric.Middleware(), &middleware.Recovery{})
		serveMux.Handle("/metrics", metric.Handler())
		http.ListenAndServe(":8080", serveMux)

		return nil
	},
	&ServeOptions{},
)

func loggerMiddleware() *middleware.Log {
	return &middleware.Log{
		AfterRequest: func(ctx context.Context, r *http.Request, observed *handler.Observation) {
			statusType := observed.Response.Status / 100

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
