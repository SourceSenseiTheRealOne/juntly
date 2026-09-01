package bookings

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceCreatesProposalBookingWithBoundedIdempotencyKey(t *testing.T) {
	t.Parallel()
	actorID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	proposalID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	store := &recordingStore{booking: Booking{ID: uuid.New(), CustomerID: actorID, State: StatePendingProviderConfirmation, Revision: 1}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actorID}}, store)

	value, err := service.Create(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, CreateBooking{SourceType: SourceProposal, SourceID: &proposalID, IdempotencyKey: "booking-request-001", ScheduledAt: time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC), PrivateLocation: "Rua privada 1"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if value.ID != store.booking.ID || store.actorID != actorID || store.created.IdempotencyKey != "booking-request-001" {
		t.Fatalf("booking/store = %#v/%s/%#v", value, store.actorID, store.created)
	}
}

func TestServiceRejectsInvalidTransitionBeforeStore(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)
	_, err := service.Transition(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, uuid.New(), Transition{ExpectedState: StatePendingProviderConfirmation, TargetState: StateCompleted, Revision: 1})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Transition() error = %v, want ErrInvalid", err)
	}
	if store.transitioned.Revision != 0 {
		t.Fatal("store called for invalid transition")
	}
}

func TestServiceAllowsProviderConfirmationTransition(t *testing.T) {
	t.Parallel()
	actorID := uuid.New()
	bookingID := uuid.New()
	store := &recordingStore{booking: Booking{ID: bookingID, State: StateConfirmed, Revision: 2}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actorID}}, store)
	value, err := service.Transition(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, bookingID, Transition{ExpectedState: StatePendingProviderConfirmation, TargetState: StateConfirmed, Revision: 1})
	if err != nil || value.State != StateConfirmed || store.actorID != actorID {
		t.Fatalf("Transition() = %#v/%v/%s", value, err, store.actorID)
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
	booking      Booking
	actorID      uuid.UUID
	created      CreateBooking
	transitioned Transition
}

func (s *recordingStore) Create(_ context.Context, actor uuid.UUID, input CreateBooking) (Booking, error) {
	s.actorID, s.created = actor, input
	return s.booking, nil
}
func (s *recordingStore) List(context.Context, uuid.UUID) ([]Booking, error) { return nil, nil }
func (s *recordingStore) Get(context.Context, uuid.UUID, uuid.UUID) (Booking, error) {
	return s.booking, nil
}
func (s *recordingStore) Transition(_ context.Context, actor, id uuid.UUID, input Transition) (Booking, error) {
	s.actorID, s.transitioned = actor, input
	return s.booking, nil
}
