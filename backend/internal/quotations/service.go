package quotations

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}
type Store interface {
	CreateRequest(context.Context, uuid.UUID, CreateRequest) (Request, error)
	ListCustomerRequests(context.Context, uuid.UUID) ([]Request, error)
	ListOpportunities(context.Context, uuid.UUID) ([]Request, error)
	SubmitProposal(context.Context, uuid.UUID, uuid.UUID, SubmitProposal) (Proposal, error)
	ListProposals(context.Context, uuid.UUID, uuid.UUID) ([]Proposal, error)
	AcceptProposal(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Proposal, error)
}
type Service interface {
	CreateRequest(context.Context, users.VerifiedIdentity, CreateRequest) (Request, error)
	ListCustomerRequests(context.Context, users.VerifiedIdentity) ([]Request, error)
	ListOpportunities(context.Context, users.VerifiedIdentity) ([]Request, error)
	SubmitProposal(context.Context, users.VerifiedIdentity, uuid.UUID, SubmitProposal) (Proposal, error)
	ListProposals(context.Context, users.VerifiedIdentity, uuid.UUID) ([]Proposal, error)
	AcceptProposal(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) (Proposal, error)
}
type service struct {
	identities IdentityReconciler
	store      Store
	now        func() time.Time
}

func NewService(identities IdentityReconciler, store Store, now func() time.Time) Service {
	return service{identities: identities, store: store, now: now}
}
func (s service) CreateRequest(ctx context.Context, identity users.VerifiedIdentity, input CreateRequest) (Request, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	now := s.currentTime()
	if utf8.RuneCountInString(input.Title) < 5 || utf8.RuneCountInString(input.Title) > 140 || utf8.RuneCountInString(input.Description) < 20 || utf8.RuneCountInString(input.Description) > 4000 || input.CategoryID == uuid.Nil || input.LocalityID == uuid.Nil || !input.ProposalDeadline.After(now) || input.ProposalDeadline.After(now.Add(90*24*time.Hour)) || (input.BudgetMinor != nil && *input.BudgetMinor <= 0) {
		return Request{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Request{}, err
	}
	v, err := s.store.CreateRequest(ctx, actor, input)
	return v, normalize(err)
}
func (s service) ListCustomerRequests(ctx context.Context, identity users.VerifiedIdentity) ([]Request, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.ListCustomerRequests(ctx, actor)
	return v, normalize(err)
}
func (s service) ListOpportunities(ctx context.Context, identity users.VerifiedIdentity) ([]Request, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.ListOpportunities(ctx, actor)
	return v, normalize(err)
}
func (s service) SubmitProposal(ctx context.Context, identity users.VerifiedIdentity, requestID uuid.UUID, input SubmitProposal) (Proposal, error) {
	input.Message = strings.TrimSpace(input.Message)
	now := s.currentTime()
	if requestID == uuid.Nil || input.PriceMinor <= 0 || utf8.RuneCountInString(input.Message) < 5 || utf8.RuneCountInString(input.Message) > 2000 || input.AvailableAt.Before(now) || (input.EstimatedMinutes != nil && (*input.EstimatedMinutes < 15 || *input.EstimatedMinutes > 525600)) || (input.ExpiresAt != nil && (!input.ExpiresAt.After(now) || input.ExpiresAt.After(now.Add(90*24*time.Hour)))) {
		return Proposal{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Proposal{}, err
	}
	v, err := s.store.SubmitProposal(ctx, actor, requestID, input)
	return v, normalize(err)
}
func (s service) ListProposals(ctx context.Context, identity users.VerifiedIdentity, requestID uuid.UUID) ([]Proposal, error) {
	if requestID == uuid.Nil {
		return nil, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	v, err := s.store.ListProposals(ctx, actor, requestID)
	return v, normalize(err)
}
func (s service) AcceptProposal(ctx context.Context, identity users.VerifiedIdentity, requestID, proposalID uuid.UUID) (Proposal, error) {
	if requestID == uuid.Nil || proposalID == uuid.Nil {
		return Proposal{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Proposal{}, err
	}
	v, err := s.store.AcceptProposal(ctx, actor, requestID, proposalID)
	return v, normalize(err)
}
func (s service) actor(ctx context.Context, identity users.VerifiedIdentity) (uuid.UUID, error) {
	if s.identities == nil || s.store == nil {
		return uuid.Nil, ErrUnavailable
	}
	u, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return uuid.Nil, ErrUnauthorized
		}
		return uuid.Nil, ErrUnavailable
	}
	if u.ID == uuid.Nil {
		return uuid.Nil, ErrUnauthorized
	}
	return u.ID, nil
}
func (s service) currentTime() time.Time {
	if s.now == nil {
		return time.Time{}
	}
	return s.now().UTC()
}
func normalize(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) {
		return err
	}
	return ErrUnavailable
}
