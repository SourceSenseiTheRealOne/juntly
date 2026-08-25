package listings

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceRejectsUnauthorizedProviderBeforeRepository(t *testing.T) {
	t.Parallel()
	for _, authorizationError := range []error{provideraccess.ErrUnauthorized, provideraccess.ErrForbidden, provideraccess.ErrUnavailable} {
		repository := &recordingRepository{}
		_, err := NewService(&recordingAuthorizer{err: authorizationError}, repository).Create(context.Background(), users.VerifiedIdentity{}, validCreate())
		if !errors.Is(err, authorizationError) {
			t.Fatalf("error = %v, want %v", err, authorizationError)
		}
		if repository.calls != 0 {
			t.Fatalf("repository calls = %d, want 0", repository.calls)
		}
	}
}

func TestServiceRejectsInvalidDraftBeforeRepository(t *testing.T) {
	t.Parallel()
	valid := validCreate()
	for name, input := range map[string]CreateListing{
		"short title":            replaceCreate(valid, func(value *CreateListing) { value.Title = "A" }),
		"short description":      replaceCreate(valid, func(value *CreateListing) { value.Description = "short" }),
		"missing category":       replaceCreate(valid, func(value *CreateListing) { value.CategoryID = uuid.Nil }),
		"missing locality":       replaceCreate(valid, func(value *CreateListing) { value.PrimaryLocalityID = uuid.Nil }),
		"unsupported price type": replaceCreate(valid, func(value *CreateListing) { value.PriceType = "subscription" }),
		"missing fixed price":    replaceCreate(valid, func(value *CreateListing) { value.PriceMinor = nil }),
		"price on quote":         replaceCreate(valid, func(value *CreateListing) { value.PriceType = PriceTypeQuote }),
		"wrong currency":         replaceCreate(valid, func(value *CreateListing) { value.Currency = "USD" }),
		"no service mode":        replaceCreate(valid, func(value *CreateListing) { value.TravelsToCustomer = false }),
	} {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			repository := &recordingRepository{}
			_, err := NewService(&recordingAuthorizer{owner: owner()}, repository).Create(context.Background(), users.VerifiedIdentity{Subject: "provider"}, input)
			if !errors.Is(err, ErrInvalidListing) {
				t.Fatalf("error = %v, want ErrInvalidListing", err)
			}
			if repository.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", repository.calls)
			}
		})
	}
}

func TestServiceCreatesAndReadsOwnerDraft(t *testing.T) {
	t.Parallel()
	input := validCreate()
	stored := listingFromCreate(input)
	repository := &recordingRepository{created: stored, found: &stored}
	service := NewService(&recordingAuthorizer{owner: owner()}, repository)

	created, err := service.Create(context.Background(), users.VerifiedIdentity{Subject: "provider"}, input)
	if err != nil || !reflect.DeepEqual(created, stored) {
		t.Fatalf("created = %#v, err = %v", created, err)
	}
	if repository.owner != owner().ID || !reflect.DeepEqual(repository.input, input) {
		t.Fatalf("repository owner/input = %s/%#v", repository.owner, repository.input)
	}
	found, err := service.Get(context.Background(), users.VerifiedIdentity{Subject: "provider"}, stored.ID)
	if err != nil || !reflect.DeepEqual(found, &stored) {
		t.Fatalf("found = %#v, err = %v", found, err)
	}
}

func TestServiceMapsRepositoryFailureToUnavailable(t *testing.T) {
	t.Parallel()
	_, err := NewService(&recordingAuthorizer{owner: owner()}, &recordingRepository{err: errors.New("private database detail")}).List(context.Background(), users.VerifiedIdentity{Subject: "provider"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
}

func TestServiceReplacesAuthorizedDraftWithExpectedRevision(t *testing.T) {
	t.Parallel()
	input := validCreate()
	stored := listingFromCreate(input)
	stored.Revision = 2
	repository := &recordingRepository{updated: stored}

	updated, err := NewService(&recordingAuthorizer{owner: owner()}, repository).ReplaceDraft(
		context.Background(), users.VerifiedIdentity{Subject: "provider"}, stored.ID, 1, input,
	)
	if err != nil || !reflect.DeepEqual(updated, stored) {
		t.Fatalf("updated = %#v, err = %v", updated, err)
	}
	if repository.owner != owner().ID || repository.id != stored.ID || repository.revision != 1 || !reflect.DeepEqual(repository.input, input) {
		t.Fatalf("repository owner/id/revision/input = %s/%s/%d/%#v", repository.owner, repository.id, repository.revision, repository.input)
	}
}

func validCreate() CreateListing {
	price := 5000
	return CreateListing{
		CategoryID:        uuid.MustParse("aaaaaaaa-aaaa-daaa-0aaa-aaaaaaaaaaaa"),
		PrimaryLocalityID: uuid.MustParse("bbbbbbbb-bbbb-dbbb-0bbb-bbbbbbbbbbbb"),
		Title:             "Reparação local de canalização",
		Description:       "Serviço local para pequenas reparações de canalização e manutenção doméstica.",
		PriceType:         PriceTypeFixed,
		PriceMinor:        &price,
		Currency:          "EUR",
		TravelsToCustomer: true,
		ReceivesCustomer:  false,
		RemoteServices:    false,
	}
}
func replaceCreate(value CreateListing, mutate func(*CreateListing)) CreateListing {
	mutate(&value)
	return value
}
func owner() users.InternalUser {
	return users.InternalUser{ID: uuid.MustParse("cccccccc-cccc-dccc-0ccc-cccccccccccc")}
}
func listingFromCreate(value CreateListing) Listing {
	return Listing{ID: uuid.MustParse("dddddddd-dddd-dddd-0ddd-dddddddddddd"), CreateListing: value, State: StateDraft, Revision: 1, CreatedAt: time.Date(2026, 8, 24, 11, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 24, 11, 0, 0, 0, time.UTC)}
}

type recordingAuthorizer struct {
	owner users.InternalUser
	err   error
}

func (a *recordingAuthorizer) RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.owner, a.err
}

type recordingRepository struct {
	owner    uuid.UUID
	id       uuid.UUID
	revision int
	input    CreateListing
	created  Listing
	updated  Listing
	found    *Listing
	err      error
	calls    int
}

func (r *recordingRepository) Create(_ context.Context, owner uuid.UUID, input CreateListing) (Listing, error) {
	r.calls++
	r.owner = owner
	r.input = input
	return r.created, r.err
}
func (r *recordingRepository) FindByOwner(_ context.Context, owner, id uuid.UUID) (*Listing, error) {
	r.calls++
	r.owner = owner
	return r.found, r.err
}
func (r *recordingRepository) ListByOwner(_ context.Context, owner uuid.UUID) ([]Listing, error) {
	r.calls++
	r.owner = owner
	return nil, r.err
}
func (r *recordingRepository) ReplaceDraft(_ context.Context, owner, id uuid.UUID, revision int, input CreateListing) (Listing, error) {
	r.calls++
	r.owner, r.id, r.revision, r.input = owner, id, revision, input
	return r.updated, r.err
}
func (r *recordingRepository) TransitionOwned(_ context.Context, actor, id uuid.UUID, from, to State, revision int, reason *string) (Listing, error) {
	r.calls++
	r.owner, r.id, r.revision = actor, id, revision
	return Listing{ID: id, State: to, Revision: revision + 1}, r.err
}
func (r *recordingRepository) TransitionModerated(_ context.Context, actor, id uuid.UUID, from, to State, revision int, reason *string) (Listing, error) {
	r.calls++
	r.owner, r.id, r.revision = actor, id, revision
	return Listing{ID: id, State: to, Revision: revision + 1}, r.err
}
