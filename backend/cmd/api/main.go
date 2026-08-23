package main

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var version = "0.1.0"

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := loadRuntimeConfig(os.Getenv)
	if err != nil {
		return err
	}
	handler, closer, err := newAPIHandler(config)
	if err != nil {
		return err
	}
	defer func() {
		_ = closer.Close()
	}()

	addr := os.Getenv("JUNTLY_API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
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

func newAPIHandler(config runtimeConfig) (http.Handler, io.Closer, error) {
	database, err := sql.Open("pgx", config.databaseURL)
	if err != nil {
		return nil, nil, err
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))

	healthService := health.NewService(version, time.Now)
	userService := users.NewService(users.NewEntRepository(client))
	accountService := accounts.NewService(userService, accounts.NewEntRepository(client))
	referenceRepository := reference.NewSQLRepository(database)
	referenceService := reference.NewService(referenceRepository)
	providerAuthorizer := provideraccess.NewService(userService, accountService)
	providerService := providers.NewService(providerAuthorizer, providers.NewEntRepository(client), referenceRepository)
	return httpapi.NewRouter(healthService, config.verifier, userService, accountService, referenceService, providerService), client, nil
}
