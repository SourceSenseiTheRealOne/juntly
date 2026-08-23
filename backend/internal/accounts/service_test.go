package accounts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceRejectsInvalidVerifiedIdentityBeforeAccountRepositoryAccess(t *testing.T) {
	t.Parallel()

	reconciler := &recordingReconciler{err: users.ErrInvalidIdentity}
	repository := &recordingRepository{}

	_, err := NewService(reconciler, repository).Get(context.Background(), users.VerifiedIdentity{})

	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("error = %v, want ErrInvalidIdentity", err)
	}
	if repository.findCalls != 0 || repository.createCalls != 0 || repository.updateCalls != 0 {
		t.Fatalf("repository calls = find:%d create:%d update:%d, want none", repository.findCalls, repository.createCalls, repository.updateCalls)
	}
}

func TestServiceGetCreatesFirstAccountWithImplicitCustomerCapability(t *testing.T) {
	t.Parallel()

	internalUser := testInternalUser()
	record := testRecord(internalUser.ID, false)
	repository := &recordingRepository{createResult: record}

	account, err := NewService(&recordingReconciler{user: internalUser}, repository).Get(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_123"},
	)

	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	assertAccountMatchesRecord(t, account, record)
	if !account.CustomerEnabled {
		t.Fatal("customer capability must always be enabled")
	}
	if repository.findCalls != 1 || repository.createCalls != 1 {
		t.Fatalf("repository calls = find:%d create:%d, want 1:1", repository.findCalls, repository.createCalls)
	}
}

func TestServiceGetReturnsExistingAccountWithoutCreate(t *testing.T) {
	t.Parallel()

	internalUser := testInternalUser()
	record := testRecord(internalUser.ID, true)
	repository := &recordingRepository{findResults: []accountFindResult{{record: record, found: true}}}

	account, err := NewService(&recordingReconciler{user: internalUser}, repository).Get(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_123"},
	)

	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	assertAccountMatchesRecord(t, account, record)
	if repository.createCalls != 0 {
		t.Fatalf("create calls = %d, want 0", repository.createCalls)
	}
}

func TestServiceGetReloadsConcurrentCreationWinner(t *testing.T) {
	t.Parallel()

	internalUser := testInternalUser()
	winner := testRecord(internalUser.ID, false)
	repository := &recordingRepository{
		findResults: []accountFindResult{{}, {record: winner, found: true}},
		createErr:   ErrAccountConflict,
	}

	account, err := NewService(&recordingReconciler{user: internalUser}, repository).Get(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_123"},
	)

	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	assertAccountMatchesRecord(t, account, winner)
	if repository.findCalls != 2 || repository.createCalls != 1 {
		t.Fatalf("repository calls = find:%d create:%d, want 2:1", repository.findCalls, repository.createCalls)
	}
}

func TestServiceSetProviderEnabledChangesCapabilityBothDirections(t *testing.T) {
	t.Parallel()

	internalUser := testInternalUser()
	initial := testRecord(internalUser.ID, false)
	enabled := initial
	enabled.ProviderEnabled = true
	disabled := enabled
	disabled.ProviderEnabled = false
	repository := &recordingRepository{
		findResults:   []accountFindResult{{record: initial, found: true}, {record: enabled, found: true}},
		updateResults: []Record{enabled, disabled},
	}
	service := NewService(&recordingReconciler{user: internalUser}, repository)
	identity := users.VerifiedIdentity{Subject: "user_123"}

	account, err := service.SetProviderEnabled(context.Background(), identity, true)
	if err != nil {
		t.Fatalf("enable provider capability: %v", err)
	}
	assertAccountMatchesRecord(t, account, enabled)

	account, err = service.SetProviderEnabled(context.Background(), identity, false)
	if err != nil {
		t.Fatalf("disable provider capability: %v", err)
	}
	assertAccountMatchesRecord(t, account, disabled)

	if repository.updateCalls != 2 || repository.updateValues[0] != true || repository.updateValues[1] != false {
		t.Fatalf("updates = calls:%d values:%v, want 2 [true false]", repository.updateCalls, repository.updateValues)
	}
}

func TestServiceMapsRepositoryFailureToUnavailable(t *testing.T) {
	t.Parallel()

	internalUser := testInternalUser()
	repository := &recordingRepository{findResults: []accountFindResult{{err: errors.New("database unavailable")}}}

	_, err := NewService(&recordingReconciler{user: internalUser}, repository).Get(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_123"},
	)

	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
	if repository.createCalls != 0 || repository.updateCalls != 0 {
		t.Fatalf("mutating calls = create:%d update:%d, want none", repository.createCalls, repository.updateCalls)
	}
}

func testInternalUser() users.InternalUser {
	return users.InternalUser{
		ID:        uuid.MustParse("5809af0d-3cf5-45ac-b120-b45e23a675a4"),
		CreatedAt: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
	}
}

func testRecord(internalUserID uuid.UUID, providerEnabled bool) Record {
	return Record{
		InternalUserID:        internalUserID,
		ProviderEnabled:       providerEnabled,
		OnboardingCompletedAt: time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC),
		CreatedAt:             time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2026, 8, 23, 12, 5, 0, 0, time.UTC),
	}
}

func assertAccountMatchesRecord(t *testing.T, account Account, record Record) {
	t.Helper()
	if account != (Account{
		CustomerEnabled:       true,
		ProviderEnabled:       record.ProviderEnabled,
		OnboardingCompletedAt: record.OnboardingCompletedAt,
	}) {
		t.Fatalf("account = %#v, want record-derived account for %#v", account, record)
	}
}

type recordingReconciler struct {
	user  users.InternalUser
	err   error
	calls int
}

func (r *recordingReconciler) Reconcile(_ context.Context, _ users.VerifiedIdentity) (users.InternalUser, bool, error) {
	r.calls++
	return r.user, false, r.err
}

type accountFindResult struct {
	record Record
	found  bool
	err    error
}

type recordingRepository struct {
	findResults   []accountFindResult
	createResult  Record
	createErr     error
	updateResults []Record
	updateErr     error
	findCalls     int
	createCalls   int
	updateCalls   int
	updateValues  []bool
}

func (r *recordingRepository) FindByInternalUserID(_ context.Context, _ uuid.UUID) (Record, bool, error) {
	r.findCalls++
	if len(r.findResults) == 0 {
		return Record{}, false, nil
	}
	result := r.findResults[0]
	r.findResults = r.findResults[1:]
	return result.record, result.found, result.err
}

func (r *recordingRepository) Create(_ context.Context, _ uuid.UUID) (Record, error) {
	r.createCalls++
	return r.createResult, r.createErr
}

func (r *recordingRepository) SetProviderEnabled(_ context.Context, _ uuid.UUID, enabled bool) (Record, error) {
	r.updateCalls++
	r.updateValues = append(r.updateValues, enabled)
	if r.updateErr != nil {
		return Record{}, r.updateErr
	}
	if len(r.updateResults) == 0 {
		return Record{}, nil
	}
	result := r.updateResults[0]
	r.updateResults = r.updateResults[1:]
	return result, nil
}
