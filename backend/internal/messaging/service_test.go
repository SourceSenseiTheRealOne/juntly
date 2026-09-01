package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceStartsListingConversationAndCreatesNotification(t *testing.T) {
	t.Parallel()
	actorID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	listingID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	store := &recordingStore{conversation: Conversation{
		ID: uuid.MustParse("33333333-3333-4333-8333-333333333333"), ListingID: &listingID,
		CustomerID: actorID, ProviderID: uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		CreatedAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
	}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actorID}}, store)

	conversation, err := service.Start(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, listingID)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if conversation.ID != store.conversation.ID || store.startedBy != actorID || store.listingID != listingID {
		t.Fatalf("Start() conversation/store = %#v/%s/%s", conversation, store.startedBy, store.listingID)
	}
}

func TestServiceRejectsBlankMessageBeforeStore(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)

	_, err := service.Send(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, uuid.New(), "   ")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Send() error = %v, want ErrInvalid", err)
	}
	if store.sentBody != "" {
		t.Fatalf("store called with %q", store.sentBody)
	}
}

func TestServiceMapsParticipantDenialToForbidden(t *testing.T) {
	t.Parallel()
	store := &recordingStore{sendErr: ErrForbidden}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)

	_, err := service.Send(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, uuid.New(), "Preciso de ajuda amanhã.")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("Send() error = %v, want ErrForbidden", err)
	}
}

func TestServiceListsOnlyReconciledParticipants(t *testing.T) {
	t.Parallel()
	actorID := uuid.New()
	store := &recordingStore{conversations: []Conversation{{ID: uuid.New(), CustomerID: actorID, ProviderID: uuid.New()}}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actorID}}, store)

	items, err := service.List(context.Background(), users.VerifiedIdentity{Subject: "user_customer"})
	if err != nil || len(items) != 1 || store.listedBy != actorID {
		t.Fatalf("List() items/error/actor = %#v/%v/%s", items, err, store.listedBy)
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
	conversation  Conversation
	conversations []Conversation
	startedBy     uuid.UUID
	listingID     uuid.UUID
	listedBy      uuid.UUID
	sentBody      string
	sendErr       error
}

func (s *recordingStore) Start(_ context.Context, actorID, listingID uuid.UUID) (Conversation, error) {
	s.startedBy, s.listingID = actorID, listingID
	return s.conversation, nil
}
func (s *recordingStore) List(_ context.Context, actorID uuid.UUID) ([]Conversation, error) {
	s.listedBy = actorID
	return s.conversations, nil
}
func (s *recordingStore) ListMessages(context.Context, uuid.UUID, uuid.UUID) ([]Message, error) {
	return nil, nil
}
func (s *recordingStore) Send(_ context.Context, _ uuid.UUID, _ uuid.UUID, body string) (Message, error) {
	s.sentBody = body
	return Message{ID: uuid.New(), Body: body}, s.sendErr
}
func (s *recordingStore) SetBlocked(context.Context, uuid.UUID, uuid.UUID, bool) error { return nil }
func (s *recordingStore) Report(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, string) error {
	return nil
}
func (s *recordingStore) Preferences(context.Context, uuid.UUID) (NotificationPreferences, error) {
	return NotificationPreferences{}, nil
}
func (s *recordingStore) ReplacePreferences(context.Context, uuid.UUID, NotificationPreferences) (NotificationPreferences, error) {
	return NotificationPreferences{}, nil
}
func (s *recordingStore) Notifications(context.Context, uuid.UUID) ([]Notification, error) {
	return nil, nil
}
func (s *recordingStore) MarkNotificationRead(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
