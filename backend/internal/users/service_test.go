package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReconcileRejectsInvalidSubjectBeforeRepositoryAccess(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{}
	_, _, err := NewService(repository).Reconcile(context.Background(), VerifiedIdentity{})

	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("error = %v, want ErrInvalidIdentity", err)
	}
	if repository.findCalls != 0 || repository.createCalls != 0 {
		t.Fatalf("repository calls = find:%d create:%d, want none", repository.findCalls, repository.createCalls)
	}
}

func TestReconcileCreatesFirstMapping(t *testing.T) {
	t.Parallel()

	created := InternalUser{
		ID:        uuid.MustParse("7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b"),
		CreatedAt: time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
	}
	repository := &recordingRepository{createResult: created}

	user, wasCreated, err := NewService(repository).Reconcile(context.Background(), VerifiedIdentity{Subject: "user_123"})

	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !wasCreated {
		t.Fatal("wasCreated = false, want true")
	}
	if user != created {
		t.Fatalf("user = %#v, want %#v", user, created)
	}
	if repository.findCalls != 1 || repository.createCalls != 1 {
		t.Fatalf("repository calls = find:%d create:%d, want 1:1", repository.findCalls, repository.createCalls)
	}
}

func TestReconcileReturnsExistingMappingWithoutCreate(t *testing.T) {
	t.Parallel()

	existing := InternalUser{
		ID:        uuid.MustParse("7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b"),
		CreatedAt: time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
	}
	repository := &recordingRepository{findResults: []findResult{{user: existing, found: true}}}

	user, wasCreated, err := NewService(repository).Reconcile(context.Background(), VerifiedIdentity{Subject: "user_123"})

	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if wasCreated {
		t.Fatal("wasCreated = true, want false")
	}
	if user != existing {
		t.Fatalf("user = %#v, want %#v", user, existing)
	}
	if repository.findCalls != 1 || repository.createCalls != 0 {
		t.Fatalf("repository calls = find:%d create:%d, want 1:0", repository.findCalls, repository.createCalls)
	}
}

func TestReconcileReloadsWinnerAfterUniqueConflict(t *testing.T) {
	t.Parallel()

	winner := InternalUser{
		ID:        uuid.MustParse("7b7b7d7e-38f9-4f0c-8a10-0fce9cf6f82b"),
		CreatedAt: time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC),
	}
	repository := &recordingRepository{
		findResults: []findResult{{}, {user: winner, found: true}},
		createErr:   ErrSubjectConflict,
	}

	user, wasCreated, err := NewService(repository).Reconcile(context.Background(), VerifiedIdentity{Subject: "user_123"})

	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if wasCreated {
		t.Fatal("wasCreated = true, want false")
	}
	if user != winner {
		t.Fatalf("user = %#v, want %#v", user, winner)
	}
	if repository.findCalls != 2 || repository.createCalls != 1 {
		t.Fatalf("repository calls = find:%d create:%d, want 2:1", repository.findCalls, repository.createCalls)
	}
}

func TestReconcileReturnsControlledFailureForRepositoryError(t *testing.T) {
	t.Parallel()

	repository := &recordingRepository{findResults: []findResult{{err: errors.New("database unavailable")}}}

	_, _, err := NewService(repository).Reconcile(context.Background(), VerifiedIdentity{Subject: "user_123"})

	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
	if repository.createCalls != 0 {
		t.Fatalf("create calls = %d, want 0", repository.createCalls)
	}
}

type findResult struct {
	user  InternalUser
	found bool
	err   error
}

type recordingRepository struct {
	findResults []findResult
	createResult InternalUser
	createErr    error
	findCalls    int
	createCalls  int
}

func (r *recordingRepository) FindBySubject(_ context.Context, _ string) (InternalUser, bool, error) {
	r.findCalls++
	if len(r.findResults) == 0 {
		return InternalUser{}, false, nil
	}

	result := r.findResults[0]
	r.findResults = r.findResults[1:]
	return result.user, result.found, result.err
}

func (r *recordingRepository) Create(_ context.Context, _ string) (InternalUser, error) {
	r.createCalls++
	return r.createResult, r.createErr
}
