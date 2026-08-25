package contactreveal

import (
	"context"
	"errors"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("contact reveal unauthorized")
	ErrForbidden    = errors.New("contact reveal forbidden")
	ErrUnavailable  = errors.New("contact reveal unavailable")
)

type Channel string

const (
	ChannelPhone    Channel = "phone"
	ChannelWhatsApp Channel = "whatsapp"
)

type IdentityReconciler interface {
	Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error)
}

type RevealStore interface {
	AuthorizeAndReserve(context.Context, uuid.UUID, uuid.UUID, Channel, time.Time) (SealedContact, error)
}

type RevealService interface {
	Reveal(context.Context, users.VerifiedIdentity, uuid.UUID, Channel) (RevealedContact, error)
}

type RevealedContact struct {
	Channel Channel
	Value   string
}

type revealService struct {
	identities IdentityReconciler
	store      RevealStore
	cipher     Cipher
	now        func() time.Time
}

func NewRevealService(identities IdentityReconciler, store RevealStore, cipher Cipher, now func() time.Time) RevealService {
	return revealService{identities: identities, store: store, cipher: cipher, now: now}
}

func (s revealService) Reveal(ctx context.Context, identity users.VerifiedIdentity, listingID uuid.UUID, channel Channel) (RevealedContact, error) {
	if listingID == uuid.Nil || !validChannel(channel) {
		return RevealedContact{}, ErrForbidden
	}
	if s.identities == nil || s.store == nil || s.cipher == nil || s.now == nil {
		return RevealedContact{}, ErrUnavailable
	}
	customer, _, err := s.identities.Reconcile(ctx, identity)
	if err != nil {
		if errors.Is(err, users.ErrInvalidIdentity) {
			return RevealedContact{}, ErrUnauthorized
		}
		return RevealedContact{}, ErrUnavailable
	}
	sealed, err := s.store.AuthorizeAndReserve(ctx, customer.ID, listingID, channel, utcDay(s.now()))
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return RevealedContact{}, ErrForbidden
		}
		return RevealedContact{}, ErrUnavailable
	}
	value, err := s.cipher.Decrypt(sealed)
	if err != nil || len(value) == 0 {
		return RevealedContact{}, ErrUnavailable
	}
	return RevealedContact{Channel: channel, Value: string(value)}, nil
}

func validChannel(channel Channel) bool {
	return channel == ChannelPhone || channel == ChannelWhatsApp
}

func utcDay(now time.Time) time.Time {
	now = now.UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
