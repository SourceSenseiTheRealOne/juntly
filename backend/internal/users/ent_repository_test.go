package users

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/internaluser"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestEntRepositoryFindsAndCreatesByExactSubject(t *testing.T) {
	client := openIntegrationClient(t)
	repository := NewEntRepository(client)
	ctx := context.Background()
	subject := "test_" + uuid.NewString()

	created, err := repository.Create(ctx, subject)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		if err := client.InternalUser.DeleteOneID(created.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup internal user: %v", err)
		}
	})

	found, exists, err := repository.FindBySubject(ctx, subject)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if !exists {
		t.Fatal("exists = false, want true")
	}
	if found != created {
		t.Fatalf("found = %#v, want %#v", found, created)
	}

	_, exists, err = repository.FindBySubject(ctx, subject+"_other")
	if err != nil {
		t.Fatalf("find missing: %v", err)
	}
	if exists {
		t.Fatal("exists = true for a different subject")
	}
}

func TestReconcileConcurrentSameSubjectProducesOneStableRow(t *testing.T) {
	client := openIntegrationClient(t)
	repository := NewEntRepository(client)
	service := NewService(repository)
	ctx := context.Background()
	subject := "test_" + uuid.NewString()

	const attempts = 8
	start := make(chan struct{})
	results := make(chan InternalUser, attempts)
	errs := make(chan error, attempts)
	var workers sync.WaitGroup
	for range attempts {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			user, _, err := service.Reconcile(ctx, VerifiedIdentity{Subject: subject})
			if err != nil {
				errs <- err
				return
			}
			results <- user
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	close(errs)

	for err := range errs {
		t.Fatalf("reconcile: %v", err)
	}

	var stable InternalUser
	for user := range results {
		if stable.ID == uuid.Nil {
			stable = user
			continue
		}
		if user != stable {
			t.Fatalf("user = %#v, want stable %#v", user, stable)
		}
	}
	if stable.ID == uuid.Nil {
		t.Fatal("no reconciliation result")
	}
	t.Cleanup(func() {
		if _, err := client.InternalUser.Delete().Where(internaluser.IDEQ(stable.ID)).Exec(ctx); err != nil {
			t.Errorf("cleanup internal user: %v", err)
		}
	})

	count, err := client.InternalUser.Query().Where(internaluser.ClerkSubjectEQ(subject)).Count(ctx)
	if err != nil {
		t.Fatalf("count subject mappings: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func openIntegrationClient(t *testing.T) *ent.Client {
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
