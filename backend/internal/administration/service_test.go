package administration

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceReadsBoundedMetricsForAdministrator(t *testing.T) {
	t.Parallel()
	actor := uuid.New()
	store := &recordingStore{metrics: Metrics{Users: 12, ActiveListings: 5}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actor}}, store)
	value, err := service.Metrics(context.Background(), users.VerifiedIdentity{Subject: "admin"})
	if err != nil || value.Users != 12 || store.actor != actor {
		t.Fatalf("Metrics() = %#v/%v/%s", value, err, store.actor)
	}
}
func TestServiceRejectsInvalidModerationActionBeforeStore(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)
	err := service.Moderate(context.Background(), users.VerifiedIdentity{Subject: "admin"}, ModerationAction{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
	if store.actor != uuid.Nil {
		t.Fatal("store called")
	}
}

type staticIdentity struct {
	user users.InternalUser
	err  error
}

func (s staticIdentity) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	return s.user, false, s.err
}

type recordingStore struct {
	metrics Metrics
	actor   uuid.UUID
}

func (s *recordingStore) Metrics(_ context.Context, actor uuid.UUID) (Metrics, error) {
	s.actor = actor
	return s.metrics, nil
}
func (s *recordingStore) Queue(_ context.Context, actor uuid.UUID) (Queue, error) {
	s.actor = actor
	return Queue{}, nil
}
func (s *recordingStore) Moderate(_ context.Context, actor uuid.UUID, _ ModerationAction) error {
	s.actor = actor
	return nil
}
