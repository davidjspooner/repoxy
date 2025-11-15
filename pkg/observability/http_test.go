package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidjspooner/go-http-server/pkg/handler"
)

func TestHTTPLoggerEnsuresRequestID(t *testing.T) {
	mw := HTTPLogger()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	handler := mw.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromRequest(r); got == "" {
			t.Fatalf("expected request id set in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got == "" {
		t.Fatalf("expected response to include %s header", RequestIDHeader)
	}
}

func TestHTTPLoggerHonorsIncomingRequestID(t *testing.T) {
	mw := HTTPLogger()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	req.Header.Set(RequestIDHeader, "abc-123")
	handler := mw.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromRequest(r); got != "abc-123" {
			t.Fatalf("expected propagated id, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != "abc-123" {
		t.Fatalf("expected response header to echo incoming id, got %q", got)
	}
}

func TestRequestIDFromRequestFallsBackToObservation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	observation, req := handler.GetObservation(req)
	observation.Request.ID = "obs-1"

	if got := RequestIDFromRequest(req); got != "obs-1" {
		t.Fatalf("expected observation id, got %q", got)
	}
}
