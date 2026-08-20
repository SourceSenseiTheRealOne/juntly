package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
)

var version = "0.1.0"

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	addr := os.Getenv("JUNTLY_API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	service := health.NewService(version, time.Now)
	server := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewRouter(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", addr)
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errs
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
