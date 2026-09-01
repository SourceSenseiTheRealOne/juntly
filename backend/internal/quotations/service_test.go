package quotations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceCreatesCustomerRequestWithReconciledOwner(t *testing.T) {
	t.Parallel()
	ownerID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	categoryID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	localityID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	deadline := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	store := &recordingStore{request: Request{ID: uuid.New(), CustomerID: ownerID}}
	service := NewService(staticIdentity{user: users.InternalUser{ID: ownerID}}, store, func() time.Time { return time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC) })

	value, err := service.CreateRequest(context.Background(), users.VerifiedIdentity{Subject: "user_customer"}, CreateRequest{Title: "Reparar telhado", Description: "Preciso de reparar uma pequena infiltração no telhado.", CategoryID: categoryID, LocalityID: localityID, ProposalDeadline: deadline})
	if err != nil {
		t.Fatalf("CreateRequest() error = %v", err)
	}
	if value.ID != store.request.ID || store.actorID != ownerID || store.created.Title != "Reparar telhado" {
		t.Fatalf("request/store = %#v/%s/%#v", value, store.actorID, store.created)
	}
}

func TestServiceRejectsInvalidProposalBeforeStore(t *testing.T) {
	t.Parallel()
	store := &recordingStore{}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store, time.Now)
	_, err := service.SubmitProposal(context.Background(), users.VerifiedIdentity{Subject: "user_provider"}, uuid.New(), SubmitProposal{PriceMinor: 0, Message: "ok"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("SubmitProposal() error = %v, want ErrInvalid", err)
	}
	if store.submitted.PriceMinor != 0 || store.actorID != uuid.Nil {
		t.Fatal("store was called for invalid proposal")
	}
}

func TestServicePreservesProposalPrivacyDenial(t *testing.T) {
	t.Parallel()
	store := &recordingStore{listErr: ErrForbidden}
	service := NewService(staticIdentity{user: users.InternalUser{ID: uuid.New()}}, store, time.Now)
	_, err := service.ListProposals(context.Background(), users.VerifiedIdentity{Subject: "user_competitor"}, uuid.New())
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("ListProposals() error = %v, want ErrForbidden", err)
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
	request   Request
	actorID   uuid.UUID
	created   CreateRequest
	submitted SubmitProposal
	listErr   error
}

func (s *recordingStore) CreateRequest(_ context.Context, actorID uuid.UUID, input CreateRequest) (Request, error) {
	s.actorID, s.created = actorID, input
	return s.request, nil
}
func (s *recordingStore) ListCustomerRequests(context.Context, uuid.UUID) ([]Request, error) {
	return nil, nil
}
func (s *recordingStore) ListOpportunities(context.Context, uuid.UUID) ([]Request, error) {
	return nil, nil
}
func (s *recordingStore) SubmitProposal(_ context.Context, actorID, requestID uuid.UUID, input SubmitProposal) (Proposal, error) {
	s.actorID, s.submitted = actorID, input
	return Proposal{}, nil
}
func (s *recordingStore) ListProposals(context.Context, uuid.UUID, uuid.UUID) ([]Proposal, error) {
	return nil, s.listErr
}
func (s *recordingStore) AcceptProposal(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Proposal, error) {
	return Proposal{}, nil
}
