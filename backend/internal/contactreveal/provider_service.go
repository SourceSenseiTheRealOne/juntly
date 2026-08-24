package contactreveal

import (
	"context"
	"errors"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ProviderAuthorizer interface {
	RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error)
}

type ChannelStore interface {
	Replace(context.Context, uuid.UUID, EncryptedChannel) (ChannelStatus, error)
}

type ProviderChannelService interface {
	Put(context.Context, users.VerifiedIdentity, ReplaceChannel) (ChannelStatus, error)
}

type ReplaceChannel struct {
	Channel       Channel
	Contact       string
	Enabled       bool
	RevealConsent bool
}

type EncryptedChannel struct {
	Channel       Channel
	Sealed        SealedContact
	KeyVersion    string
	Enabled       bool
	RevealConsent bool
}

type ChannelStatus struct {
	Channel       Channel
	Configured    bool
	Enabled       bool
	RevealConsent bool
}

type providerChannelService struct {
	authorizer ProviderAuthorizer
	store      ChannelStore
	cipher     Cipher
}

func NewProviderChannelService(authorizer ProviderAuthorizer, store ChannelStore, cipher Cipher) ProviderChannelService {
	return providerChannelService{authorizer: authorizer, store: store, cipher: cipher}
}

func (s providerChannelService) Put(ctx context.Context, identity users.VerifiedIdentity, input ReplaceChannel) (ChannelStatus, error) {
	if s.authorizer == nil || s.store == nil || s.cipher == nil {
		return ChannelStatus{}, ErrUnavailable
	}
	owner, err := s.authorizer.RequireProvider(ctx, identity)
	if err != nil {
		if errors.Is(err, provideraccess.ErrUnauthorized) {
			return ChannelStatus{}, ErrUnauthorized
		}
		if errors.Is(err, provideraccess.ErrForbidden) {
			return ChannelStatus{}, ErrForbidden
		}
		return ChannelStatus{}, ErrUnavailable
	}
	input.Contact = strings.TrimSpace(input.Contact)
	if !validChannel(input.Channel) || !validE164(input.Contact) {
		return ChannelStatus{}, ErrForbidden
	}
	sealed, err := s.cipher.Encrypt([]byte(input.Contact))
	if err != nil {
		return ChannelStatus{}, ErrUnavailable
	}
	status, err := s.store.Replace(ctx, owner.ID, EncryptedChannel{Channel: input.Channel, Sealed: sealed, KeyVersion: "v1", Enabled: input.Enabled, RevealConsent: input.RevealConsent})
	if err != nil {
		return ChannelStatus{}, ErrUnavailable
	}
	return status, nil
}

func validE164(value string) bool {
	if len(value) < 9 || len(value) > 16 || value[0] != '+' || value[1] == '0' {
		return false
	}
	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
