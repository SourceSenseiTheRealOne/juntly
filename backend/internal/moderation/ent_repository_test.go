package moderation

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/platformrole"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEntRepositoryReadsOnlyPersistedModeratorGrant(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	user, _, err := users.NewService(users.NewEntRepository(client)).Reconcile(ctx, users.VerifiedIdentity{Subject: "moderation_user_" + uuid.NewString()})
	if err != nil {
		t.Fatalf("reconcile user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := client.PlatformRole.Delete().Where(platformrole.InternalUserIDEQ(user.ID)).Exec(ctx); err != nil {
			t.Errorf("cleanup grant: %v", err)
		}
		if err := client.InternalUser.DeleteOneID(user.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup user: %v", err)
		}
	})
	repository := NewEntRepository(client)
	granted, err := repository.HasModeratorGrant(ctx, user.ID)
	if err != nil || granted {
		t.Fatalf("pre-grant = %v, err = %v", granted, err)
	}
	if err := client.PlatformRole.Create().SetInternalUserID(user.ID).SetRole("moderator").Exec(ctx); err != nil {
		t.Fatalf("grant moderator: %v", err)
	}
	granted, err = repository.HasModeratorGrant(ctx, user.ID)
	if err != nil || !granted {
		t.Fatalf("moderator grant = %v, err = %v", granted, err)
	}
}
