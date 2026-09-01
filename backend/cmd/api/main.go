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
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/administration"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/bookings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/discovery"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/entitlements"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/health"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/messaging"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/moderation"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/providers"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/quotations"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/readiness"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reference"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reviews"
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
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(10)
	database.SetConnMaxLifetime(30 * time.Minute)
	database.SetConnMaxIdleTime(5 * time.Minute)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))

	healthService := health.NewService(version, time.Now)
	readinessService := readiness.NewService(database)
	userService := users.NewService(users.NewEntRepository(client))
	accountService := accounts.NewService(userService, accounts.NewEntRepository(client))
	referenceRepository := reference.NewSQLRepository(database)
	referenceService := reference.NewService(referenceRepository)
	publicDiscovery := discovery.NewService(discovery.NewSQLRepository(database))
	providerAuthorizer := provideraccess.NewService(userService, accountService)
	contactChannels := contactreveal.NewProviderChannelService(providerAuthorizer, contactreveal.NewSQLChannelStore(database), config.contactCipher)
	contactReveal := contactreveal.NewRevealService(userService, contactreveal.NewSQLRevealStore(database), config.contactCipher, time.Now)
	messagingService := messaging.NewService(userService, messaging.NewSQLStore(database))
	quotationService := quotations.NewService(userService, quotations.NewSQLStore(database), time.Now)
	bookingService := bookings.NewService(userService, bookings.NewSQLStore(database))
	reviewService := reviews.NewService(userService, reviews.NewSQLStore(database))
	entitlementService := entitlements.NewService(userService, entitlements.NewSQLStore(database))
	administrationService := administration.NewService(userService, administration.NewSQLStore(database))
	providerService := providers.NewService(providerAuthorizer, providers.NewEntRepository(client), referenceRepository)
	listingRepository := listings.NewEntRepository(client)
	listingDrafts := listings.NewService(providerAuthorizer, listingRepository)
	moderatorAuthorizer := moderation.NewService(userService, moderation.NewEntRepository(client))
	listingLifecycle := listings.NewLifecycleService(providerAuthorizer, moderatorAuthorizer, listingRepository)
	listingMedia := listingmedia.NewService(providerAuthorizer, listingmedia.NewEntRepository(client), listingmedia.NewUnavailableStorage())
	ownerListings := listings.NewOwnerService(listingDrafts, listingLifecycle, listingMedia)
	moderationQueue := moderation.NewQueueService(moderatorAuthorizer, listingRepository)
	moderationReview := moderation.NewReviewService(moderationQueue, listingLifecycle)
	return httpapi.NewRouter(healthService, readinessService, config.verifier, userService, accountService, referenceService, providerService, ownerListings, moderationReview, publicDiscovery, contactChannels, contactReveal, messagingService, quotationService, bookingService, reviewService, entitlementService, administrationService), client, nil
}
