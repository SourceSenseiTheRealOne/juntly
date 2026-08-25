package accounts

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/useraccount"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEntRepositoryCreatesFindsAndUpdatesAccount(t *testing.T) {
	client := openAccountIntegrationClient(t)
	ctx := context.Background()
	identity := users.VerifiedIdentity{Subject: "test_account_" + uuid.NewString()}
	internalUser, _, err := users.NewService(users.NewEntRepository(client)).Reconcile(ctx, identity)
	if err != nil {
		t.Fatalf("reconcile internal user: %v", err)
	}
	t.Cleanup(func() {
		if err := client.InternalUser.DeleteOneID(internalUser.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup internal user: %v", err)
		}
	})

	repository := NewEntRepository(client)
	created, err := repository.Create(ctx, internalUser.ID)
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	if created.InternalUserID != internalUser.ID {
		t.Fatalf("internal user ID = %s, want %s", created.InternalUserID, internalUser.ID)
	}
	if created.ProviderEnabled {
		t.Fatal("provider capability must default to false")
	}

	found, exists, err := repository.FindByInternalUserID(ctx, internalUser.ID)
	if err != nil {
		t.Fatalf("find account: %v", err)
	}
	if !exists || found != created {
		t.Fatalf("found = %#v exists=%t, want %#v true", found, exists, created)
	}

	enabled, err := repository.SetProviderEnabled(ctx, internalUser.ID, true)
	if err != nil {
		t.Fatalf("enable provider capability: %v", err)
	}
	if !enabled.ProviderEnabled {
		t.Fatal("provider capability was not enabled")
	}
	if enabled.OnboardingCompletedAt != created.OnboardingCompletedAt {
		t.Fatalf("onboarding timestamp changed: got %s want %s", enabled.OnboardingCompletedAt, created.OnboardingCompletedAt)
	}

	disabled, err := repository.SetProviderEnabled(ctx, internalUser.ID, false)
	if err != nil {
		t.Fatalf("disable provider capability: %v", err)
	}
	if disabled.ProviderEnabled {
		t.Fatal("provider capability was not disabled")
	}
	if disabled.OnboardingCompletedAt != created.OnboardingCompletedAt {
		t.Fatalf("onboarding timestamp changed: got %s want %s", disabled.OnboardingCompletedAt, created.OnboardingCompletedAt)
	}
}

func TestConcurrentFirstAccountReadsProduceOneStableRow(t *testing.T) {
	client := openAccountIntegrationClient(t)
	ctx := context.Background()
	identity := users.VerifiedIdentity{Subject: "test_account_race_" + uuid.NewString()}
	identityService := users.NewService(users.NewEntRepository(client))
	accountService := NewService(identityService, NewEntRepository(client))

	const attempts = 8
	start := make(chan struct{})
	results := make(chan Account, attempts)
	errs := make(chan error, attempts)
	var workers sync.WaitGroup
	for range attempts {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			account, err := accountService.Get(ctx, identity)
			if err != nil {
				errs <- err
				return
			}
			results <- account
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	close(errs)

	for err := range errs {
		t.Fatalf("get account: %v", err)
	}

	var stable Account
	for account := range results {
		if stable.OnboardingCompletedAt.IsZero() {
			stable = account
			continue
		}
		if account != stable {
			t.Fatalf("account = %#v, want stable %#v", account, stable)
		}
	}
	if stable.OnboardingCompletedAt.IsZero() {
		t.Fatal("no account result")
	}
	if !stable.CustomerEnabled || stable.ProviderEnabled {
		t.Fatalf("capabilities = customer:%t provider:%t, want true:false", stable.CustomerEnabled, stable.ProviderEnabled)
	}

	internalUser, _, err := identityService.Reconcile(ctx, identity)
	if err != nil {
		t.Fatalf("reload internal user: %v", err)
	}
	t.Cleanup(func() {
		if err := client.InternalUser.DeleteOneID(internalUser.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup internal user: %v", err)
		}
	})

	count, err := client.UserAccount.Query().Where(useraccount.IDEQ(internalUser.ID)).Count(ctx)
	if err != nil {
		t.Fatalf("count account rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("account row count = %d, want 1", count)
	}
}

func openAccountIntegrationClient(t *testing.T) *ent.Client {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open pgx database: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Ent client: %v", err)
		}
	})
	return client
}
