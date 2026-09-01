package entitlements

import (
	"context"
	"errors"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	"testing"
)

func TestServiceRequestsConfiguredSubscriptionForProvider(t *testing.T) {
	t.Parallel()
	actor := uuid.New()
	plan := uuid.New()
	store := &recordingStore{subscription: Subscription{ID: uuid.New(), ProviderID: actor, PlanID: plan, Status: StatusActive}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actor}}, store)
	value, err := service.RequestSubscription(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, plan)
	if err != nil || value.PlanID != plan || store.actorID != actor {
		t.Fatalf("RequestSubscription() = %#v/%v/%s", value, err, store.actorID)
	}
}
func TestServiceRejectsMissingPromotionIdentifiers(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store)
	_, err := service.RequestPromotion(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, uuid.Nil, uuid.New())
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error = %v", err)
	}
	if store.actorID != uuid.Nil {
		t.Fatal("store called")
	}
}
func TestServiceReadsServerOwnedEntitlements(t *testing.T) {
	t.Parallel()
	actor := uuid.New()
	store := &recordingStore{access: Access{MaxActiveListings: 5, MaxPhotosPerListing: 10, AnalyticsEnabled: true}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: actor}}, store)
	value, err := service.Access(context.Background(), users.VerifiedIdentity{Subject: "user_provider"})
	if err != nil || value.MaxActiveListings != 5 || store.actorID != actor {
		t.Fatalf("Access() = %#v/%v/%s", value, err, store.actorID)
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
	subscription Subscription
	access       Access
	actorID      uuid.UUID
}

func (s *recordingStore) Catalog(context.Context) (Catalog, error) { return Catalog{}, nil }
func (s *recordingStore) RequestSubscription(_ context.Context, actor, plan uuid.UUID) (Subscription, error) {
	s.actorID = actor
	return s.subscription, nil
}
func (s *recordingStore) CurrentSubscription(context.Context, uuid.UUID) (*Subscription, error) {
	return nil, nil
}
func (s *recordingStore) RequestPromotion(_ context.Context, actor, listing, period uuid.UUID) (Promotion, error) {
	s.actorID = actor
	return Promotion{}, nil
}
func (s *recordingStore) ListPromotions(context.Context, uuid.UUID) ([]Promotion, error) {
	return nil, nil
}
func (s *recordingStore) Access(_ context.Context, actor uuid.UUID) (Access, error) {
	s.actorID = actor
	return s.access, nil
}
