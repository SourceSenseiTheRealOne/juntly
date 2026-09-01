package reviews

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceCreatesVerifiedCompletedBookingReview(t *testing.T) {
	t.Parallel()
	actor := uuid.New()
	bookingID := uuid.New()
	store := &recordingStore{review: Review{ID: uuid.New(), BookingID: bookingID, CustomerID: actor, VerifiedBooking: true}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actor}}, store)
	value, err := service.Create(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, CreateReview{BookingID: bookingID, Rating: 5, Body: "Trabalho muito bem executado."})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !value.VerifiedBooking || store.actorID != actor || store.input.Rating != 5 {
		t.Fatalf("review/store = %#v/%s/%#v", value, store.actorID, store.input)
	}
}
func TestServiceRejectsInvalidRatingBeforeStore(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)
	_, err := service.Create(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, CreateReview{BookingID: uuid.New(), Rating: 6, Body: "Texto suficientemente longo."})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
	if store.actorID != uuid.Nil {
		t.Fatal("store called")
	}
}
func TestServiceAllowsProviderResponse(t *testing.T) {
	t.Parallel()
	actor := uuid.New()
	reviewID := uuid.New()
	store := &recordingStore{review: Review{ID: reviewID, ProviderResponse: "Obrigado."}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actor}}, store)
	value, err := service.Respond(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, reviewID, "Obrigado pelo comentário.")
	if err != nil || value.ID != reviewID || store.actorID != actor {
		t.Fatalf("Respond() = %#v/%v/%s", value, err, store.actorID)
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
	review  Review
	actorID uuid.UUID
	input   CreateReview
}

func (s *recordingStore) Create(_ context.Context, actor uuid.UUID, input CreateReview) (Review, error) {
	s.actorID, s.input = actor, input
	return s.review, nil
}
func (s *recordingStore) ListForProvider(context.Context, uuid.UUID) ([]Review, error) {
	return nil, nil
}
func (s *recordingStore) Respond(_ context.Context, actor, reviewID uuid.UUID, response string) (Review, error) {
	s.actorID = actor
	return s.review, nil
}
func (s *recordingStore) Aggregate(context.Context, uuid.UUID) (Aggregate, error) {
	return Aggregate{}, nil
}
