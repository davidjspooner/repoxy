package observability

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/davidjspooner/go-http-server/pkg/handler"
	"github.com/davidjspooner/go-http-server/pkg/middleware"
)

// RequestIDHeader is the HTTP header used to propagate correlation IDs.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// HTTPLogger returns a middleware that ensures correlation IDs are set and logs structured request details.
func HTTPLogger() handler.Middleware {
	logMW := &middleware.Log{
		AfterRequest: func(ctx context.Context, r *http.Request, observed *handler.Observation) {
			attrs := []any{
				slog.String("req_id", observed.Request.ID),
				slog.String("trace_id", observed.Request.ID),
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
	return handler.MiddlewareFunc(func(next http.Handler) http.Handler {
		logged := logMW.WrapHandler(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			observed, r := handler.GetObservation(r)
			r = EnsureRequestID(r, w, observed)
			logged.ServeHTTP(w, r)
		})
	})
}

func EnsureRequestID(r *http.Request, w http.ResponseWriter, observation *handler.Observation) *http.Request {
	if observation == nil {
		var obs *handler.Observation
		obs, r = handler.GetObservation(r)
		observation = obs
	}
	requestID := ""
	if r != nil {
		requestID = strings.TrimSpace(r.Header.Get(RequestIDHeader))
	}
	if requestID == "" && observation != nil {
		requestID = observation.Request.ID
	}
	if observation != nil && requestID != "" {
		observation.Request.ID = requestID
	}
	if r != nil {
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		r = r.WithContext(ctx)
	}
	if w != nil && requestID != "" {
		w.Header().Set(RequestIDHeader, requestID)
	}
	return r
}

// RequestIDFromContext extracts the correlation ID stored by the middleware.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// RequestIDFromRequest returns the correlation ID bound to the HTTP request.
func RequestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := RequestIDFromContext(r.Context()); id != "" {
		return id
	}
	observation, _ := handler.GetObservation(r)
	if observation != nil && observation.Request.ID != "" {
		return observation.Request.ID
	}
	return ""
}

// ApplyRequestIDHeader sets the correlation ID header on outbound upstream requests.
func ApplyRequestIDHeader(req *http.Request, reqID string) {
	if req == nil {
		return
	}
	if reqID == "" {
		reqID = RequestIDFromContext(req.Context())
	}
	if reqID == "" {
		return
	}
	req.Header.Set(RequestIDHeader, reqID)
}
