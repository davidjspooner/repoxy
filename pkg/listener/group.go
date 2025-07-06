package listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type Group struct {
	Listeners []*Config `yaml:"listeners"`
}

func (lg *Group) ListenAndServe(ctx context.Context, handler http.Handler) error {
	if len(lg.Listeners) == 0 {
		return fmt.Errorf("no listeners configured")
	}

	servers := make([]*http.Server, 0, len(lg.Listeners))
	for _, listener := range lg.Listeners {
		server, err := listener.server(handler)
		if err != nil {
			return err
		}
		servers = append(servers, server)
	}

	// Channel to collect startup errors and successful starts
	errChan := make(chan error, len(servers))
	startedChan := make(chan string, len(servers))

	// Start all servers
	for _, server := range servers {
		go func(s *http.Server) {
			// Create a listener first to detect binding errors immediately
			var listener net.Listener
			var err error

			if s.TLSConfig != nil {
				listener, err = tls.Listen("tcp", s.Addr, s.TLSConfig)
			} else {
				listener, err = net.Listen("tcp", s.Addr)
			}

			if err != nil {
				errChan <- fmt.Errorf("failed to bind %s: %w", s.Addr, err)
				return
			}

			// Signal successful startup
			startedChan <- s.Addr

			// Start serving
			err = s.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				slog.Error("Server failed", "addr", s.Addr, "error", err)
			}
		}(server)
	}

	// Wait for all servers to start or first error
	started := 0
	for started < len(servers) {
		select {
		case addr := <-startedChan:
			started++
			slog.Info("Server started", "addr", addr)
		case err := <-errChan:
			return fmt.Errorf("server startup failed: %w", err)
		case <-ctx.Done():
			return fmt.Errorf("startup cancelled: %w", ctx.Err())
		}
	}

	slog.Info("All servers started successfully", "count", len(servers))

	// Wait for shutdown signal
	<-ctx.Done()
	slog.Info("Shutdown signal received")

	// Shutdown all servers concurrently
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	shutdownErrors := make(chan error, len(servers))

	for _, server := range servers {
		wg.Add(1)
		go func(s *http.Server) {
			defer wg.Done()
			if err := s.Shutdown(shutdownCtx); err != nil {
				shutdownErrors <- fmt.Errorf("failed to shutdown %s: %w", s.Addr, err)
			}
		}(server)
	}

	// Wait for all shutdowns to complete
	wg.Wait()
	close(shutdownErrors)

	// Collect any shutdown errors
	var errs ErrorList
	for err := range shutdownErrors {
		errs = append(errs, err)
	}

	if !errs.IsEmpty() {
		return errs
	}

	slog.Info("All servers shut down successfully")
	return nil
}
