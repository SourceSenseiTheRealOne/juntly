package providers

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"sync"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/locality"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/spokenlanguage"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEntRepositoryCreatesReadsReplacesAndKeepsIdempotentTimestamps(t *testing.T) {
	client := openProviderIntegrationClient(t)
	ctx := context.Background()
	owner := createProviderOwner(t, client)
	localities, languages := seededProfileReferences(t, client)
	repository := NewEntRepository(client)

	firstInput := integrationReplacement(localities[:1], languages[:1])
	created, err := repository.Replace(ctx, owner.ID, firstInput)
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	assertProfileMatchesInput(t, created, firstInput)

	found, err := repository.FindByOwner(ctx, owner.ID)
	if err != nil || found == nil {
		t.Fatalf("find profile = %#v, err = %v", found, err)
	}
	assertProfileMatchesInput(t, *found, firstInput)

	secondInput := integrationReplacement(localities[:2], languages[:2])
	secondInput.DisplayName = "Prestador atualizado"
	updated, err := repository.Replace(ctx, owner.ID, secondInput)
	if err != nil {
		t.Fatalf("replace profile: %v", err)
	}
	assertProfileMatchesInput(t, updated, secondInput)
	if updated.CreatedAt != created.CreatedAt {
		t.Fatalf("created timestamp changed: got %s want %s", updated.CreatedAt, created.CreatedAt)
	}

	repeated, err := repository.Replace(ctx, owner.ID, secondInput)
	if err != nil {
		t.Fatalf("repeat profile replacement: %v", err)
	}
	if repeated.UpdatedAt != updated.UpdatedAt {
		t.Fatalf("idempotent updated timestamp changed: got %s want %s", repeated.UpdatedAt, updated.UpdatedAt)
	}

	missing, err := repository.FindByOwner(ctx, uuid.New())
	if err != nil || missing != nil {
		t.Fatalf("cross-owner lookup = %#v, err = %v, want nil nil", missing, err)
	}
}

func TestEntRepositoryRollsBackScalarChangesWhenChildReplacementFails(t *testing.T) {
	client := openProviderIntegrationClient(t)
	ctx := context.Background()
	owner := createProviderOwner(t, client)
	localities, languages := seededProfileReferences(t, client)
	repository := NewEntRepository(client)
	originalInput := integrationReplacement(localities[:1], languages[:1])
	original, err := repository.Replace(ctx, owner.ID, originalInput)
	if err != nil {
		t.Fatalf("create original profile: %v", err)
	}

	invalid := originalInput
	invalid.DisplayName = "Must roll back"
	invalid.LanguageCodes = []string{"missing-language"}
	if _, err := repository.Replace(ctx, owner.ID, invalid); err == nil {
		t.Fatal("invalid child replacement error = nil")
	}

	found, err := repository.FindByOwner(ctx, owner.ID)
	if err != nil || found == nil {
		t.Fatalf("find after rollback = %#v, err = %v", found, err)
	}
	assertProfileMatchesInput(t, *found, originalInput)
	if found.UpdatedAt != original.UpdatedAt {
		t.Fatalf("rollback changed updated timestamp: got %s want %s", found.UpdatedAt, original.UpdatedAt)
	}
}

func TestEntRepositoryConcurrentFirstReplacementProducesOneStableProfile(t *testing.T) {
	client := openProviderIntegrationClient(t)
	ctx := context.Background()
	owner := createProviderOwner(t, client)
	localities, languages := seededProfileReferences(t, client)
	repository := NewEntRepository(client)
	input := integrationReplacement(localities[:2], languages[:2])

	const attempts = 8
	start := make(chan struct{})
	results := make(chan Profile, attempts)
	errs := make(chan error, attempts)
	var workers sync.WaitGroup
	for range attempts {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			profile, err := repository.Replace(ctx, owner.ID, input)
			if err != nil {
				errs <- err
				return
			}
			results <- profile
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent replace: %v", err)
	}

	var stable *Profile
	for profile := range results {
		if stable == nil {
			value := profile
			stable = &value
			continue
		}
		if !reflect.DeepEqual(profile, *stable) {
			t.Fatalf("profile = %#v, want stable %#v", profile, *stable)
		}
	}
	if stable == nil {
		t.Fatal("no concurrent profile result")
	}
	assertProfileMatchesInput(t, *stable, input)

	var count int
	if err := providerDatabase(t, client).QueryRowContext(ctx, "select count(*) from public.provider_profiles where internal_user_id = $1", owner.ID).Scan(&count); err != nil {
		t.Fatalf("count provider profiles: %v", err)
	}
	if count != 1 {
		t.Fatalf("provider profile count = %d, want 1", count)
	}
}

func openProviderIntegrationClient(t *testing.T) *ent.Client {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, database)))
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	providerDatabases.Store(client, database)
	return client
}

var providerDatabases sync.Map

func providerDatabase(t *testing.T, client *ent.Client) *sql.DB {
	t.Helper()
	value, ok := providerDatabases.Load(client)
	if !ok {
		t.Fatal("provider database not registered")
	}
	return value.(*sql.DB)
}

func createProviderOwner(t *testing.T, client *ent.Client) users.InternalUser {
	t.Helper()
	ctx := context.Background()
	identity := users.VerifiedIdentity{Subject: "test_provider_" + uuid.NewString()}
	owner, _, err := users.NewService(users.NewEntRepository(client)).Reconcile(ctx, identity)
	if err != nil {
		t.Fatalf("reconcile owner: %v", err)
	}
	accountRepository := accounts.NewEntRepository(client)
	if _, err := accountRepository.Create(ctx, owner.ID); err != nil {
		t.Fatalf("create owner account: %v", err)
	}
	if _, err := accountRepository.SetProviderEnabled(ctx, owner.ID, true); err != nil {
		t.Fatalf("enable provider: %v", err)
	}
	t.Cleanup(func() {
		if err := client.InternalUser.DeleteOneID(owner.ID).Exec(ctx); err != nil {
			t.Errorf("cleanup owner: %v", err)
		}
	})
	return owner
}

func seededProfileReferences(t *testing.T, client *ent.Client) ([]uuid.UUID, []string) {
	t.Helper()
	ctx := context.Background()
	localityIDs, err := client.Locality.Query().Where(locality.ActiveEQ(true)).Order(ent.Asc(locality.FieldID)).IDs(ctx)
	if err != nil || len(localityIDs) < 2 {
		t.Fatalf("seeded localities = %v, err = %v", localityIDs, err)
	}
	languageIDs, err := client.SpokenLanguage.Query().Where(spokenlanguage.ActiveEQ(true)).Order(ent.Asc(spokenlanguage.FieldID)).IDs(ctx)
	if err != nil || len(languageIDs) < 2 {
		t.Fatalf("seeded languages = %v, err = %v", languageIDs, err)
	}
	return localityIDs, languageIDs
}

func integrationReplacement(localityIDs []uuid.UUID, languageCodes []string) ReplaceProfile {
	return ReplaceProfile{
		DisplayName:         "Prestador integrado",
		ProviderType:        ProviderTypeProfessional,
		Bio:                 "Perfil sintético de integração.",
		PrimaryLocalityID:   localityIDs[0],
		ServiceLocalityIDs:  append([]uuid.UUID(nil), localityIDs...),
		MaxTravelDistanceKM: 25,
		TravelsToCustomer:   true,
		ReceivesCustomer:    false,
		RemoteServices:      false,
		LanguageCodes:       append([]string(nil), languageCodes...),
	}
}

func assertProfileMatchesInput(t *testing.T, profile Profile, input ReplaceProfile) {
	t.Helper()
	if profile.DisplayName != input.DisplayName || profile.ProviderType != input.ProviderType || profile.Bio != input.Bio || profile.PrimaryLocalityID != input.PrimaryLocalityID || profile.MaxTravelDistanceKM != input.MaxTravelDistanceKM || profile.TravelsToCustomer != input.TravelsToCustomer || profile.ReceivesCustomer != input.ReceivesCustomer || profile.RemoteServices != input.RemoteServices {
		t.Fatalf("profile scalars = %#v, want %#v", profile, input)
	}
	if !reflect.DeepEqual(profile.ServiceLocalityIDs, input.ServiceLocalityIDs) {
		t.Fatalf("service localities = %v, want %v", profile.ServiceLocalityIDs, input.ServiceLocalityIDs)
	}
	if !reflect.DeepEqual(profile.LanguageCodes, input.LanguageCodes) {
		t.Fatalf("language codes = %v, want %v", profile.LanguageCodes, input.LanguageCodes)
	}
	if profile.CreatedAt.IsZero() || profile.UpdatedAt.IsZero() {
		t.Fatalf("profile timestamps are zero: %#v", profile)
	}
}
