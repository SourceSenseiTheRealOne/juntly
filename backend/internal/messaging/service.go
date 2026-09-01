package messaging

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type Store interface {
	Start(context.Context, uuid.UUID, uuid.UUID) (Conversation, error)
	List(context.Context, uuid.UUID) ([]Conversation, error)
	ListMessages(context.Context, uuid.UUID, uuid.UUID) ([]Message, error)
	Send(context.Context, uuid.UUID, uuid.UUID, string) (Message, error)
	SetBlocked(context.Context, uuid.UUID, uuid.UUID, bool) error
	Report(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, string) error
	Preferences(context.Context, uuid.UUID) (NotificationPreferences, error)
	ReplacePreferences(context.Context, uuid.UUID, NotificationPreferences) (NotificationPreferences, error)
	Notifications(context.Context, uuid.UUID) ([]Notification, error)
	MarkNotificationRead(context.Context, uuid.UUID, uuid.UUID) error
}

type Service interface {
	Start(context.Context, users.VerifiedIdentity, uuid.UUID) (Conversation, error)
	List(context.Context, users.VerifiedIdentity) ([]Conversation, error)
	ListMessages(context.Context, users.VerifiedIdentity, uuid.UUID) ([]Message, error)
	Send(context.Context, users.VerifiedIdentity, uuid.UUID, string) (Message, error)
	SetBlocked(context.Context, users.VerifiedIdentity, uuid.UUID, bool) error
	Report(context.Context, users.VerifiedIdentity, uuid.UUID, *uuid.UUID, string) error
	Preferences(context.Context, users.VerifiedIdentity) (NotificationPreferences, error)
	ReplacePreferences(context.Context, users.VerifiedIdentity, NotificationPreferences) (NotificationPreferences, error)
	Notifications(context.Context, users.VerifiedIdentity) ([]Notification, error)
	MarkNotificationRead(context.Context, users.VerifiedIdentity, uuid.UUID) error
}

type service struct {
	identities IdentityReconciler
	store      Store
}

func NewService(identities IdentityReconciler, store Store) Service {
	return service{identities: identities, store: store}
}

func (s service) Start(ctx context.Context, identity users.VerifiedIdentity, listingID uuid.UUID) (Conversation, error) {
	if listingID == uuid.Nil {
		return Conversation{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Conversation{}, err
	}
	value, err := s.store.Start(ctx, actor, listingID)
	return value, normalize(err)
}
func (s service) List(ctx context.Context, identity users.VerifiedIdentity) ([]Conversation, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	value, err := s.store.List(ctx, actor)
	return value, normalize(err)
}
func (s service) ListMessages(ctx context.Context, identity users.VerifiedIdentity, conversationID uuid.UUID) ([]Message, error) {
	if conversationID == uuid.Nil {
		return nil, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	value, err := s.store.ListMessages(ctx, actor, conversationID)
	return value, normalize(err)
}
func (s service) Send(ctx context.Context, identity users.VerifiedIdentity, conversationID uuid.UUID, body string) (Message, error) {
	body = strings.TrimSpace(body)
	if conversationID == uuid.Nil || body == "" || utf8.RuneCountInString(body) > 4000 {
		return Message{}, ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return Message{}, err
	}
	value, err := s.store.Send(ctx, actor, conversationID, body)
	return value, normalize(err)
}
func (s service) SetBlocked(ctx context.Context, identity users.VerifiedIdentity, conversationID uuid.UUID, blocked bool) error {
	if conversationID == uuid.Nil {
		return ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return err
	}
	return normalize(s.store.SetBlocked(ctx, actor, conversationID, blocked))
}
func (s service) Report(ctx context.Context, identity users.VerifiedIdentity, conversationID uuid.UUID, messageID *uuid.UUID, reason string) error {
	reason = strings.TrimSpace(reason)
	if conversationID == uuid.Nil || utf8.RuneCountInString(reason) < 5 || utf8.RuneCountInString(reason) > 500 {
		return ErrInvalid
	}
	if messageID != nil && *messageID == uuid.Nil {
		return ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return err
	}
	return normalize(s.store.Report(ctx, actor, conversationID, messageID, reason))
}
func (s service) Preferences(ctx context.Context, identity users.VerifiedIdentity) (NotificationPreferences, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return NotificationPreferences{}, err
	}
	value, err := s.store.Preferences(ctx, actor)
	return value, normalize(err)
}
func (s service) ReplacePreferences(ctx context.Context, identity users.VerifiedIdentity, value NotificationPreferences) (NotificationPreferences, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return NotificationPreferences{}, err
	}
	updated, err := s.store.ReplacePreferences(ctx, actor, value)
	return updated, normalize(err)
}
func (s service) Notifications(ctx context.Context, identity users.VerifiedIdentity) ([]Notification, error) {
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return nil, err
	}
	items, err := s.store.Notifications(ctx, actor)
	return items, normalize(err)
}
func (s service) MarkNotificationRead(ctx context.Context, identity users.VerifiedIdentity, notificationID uuid.UUID) error {
	if notificationID == uuid.Nil {
		return ErrInvalid
	}
	actor, err := s.actor(ctx, identity)
	if err != nil {
		return err
	}
	return normalize(s.store.MarkNotificationRead(ctx, actor, notificationID))
}
func (s service) actor(ctx context.Context, identity users.VerifiedIdentity) (uuid.UUID, error) {
	if s.identities == nil || s.store == nil {
		return uuid.Nil, ErrUnavailable
	}
	user, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return uuid.Nil, ErrUnauthorized
		}
		return uuid.Nil, ErrUnavailable
	}
	if user.ID == uuid.Nil {
		return uuid.Nil, ErrUnauthorized
	}
	return user.ID, nil
}
func normalize(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden) || errors.Is(err, ErrNotFound) {
		return err
	}
	return ErrUnavailable
}
