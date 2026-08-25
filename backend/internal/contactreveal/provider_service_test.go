package contactreveal

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestProviderChannelServiceDeniesBeforeEncrypting(t *testing.T) {
	t.Parallel()
	cipher := &providerCountingCipher{}
	store := &recordingChannelStore{}
	service := NewProviderChannelService(&recordingProviderAuthorizer{err: provideraccess.ErrForbidden}, store, cipher)
	_, err := service.Put(context.Background(), users.VerifiedIdentity{Subject: "provider"}, ReplaceChannel{Channel: ChannelPhone, Contact: "+12025550123", Enabled: true, RevealConsent: true})
	if !errors.Is(err, ErrForbidden) || cipher.encryptCalls != 0 || store.calls != 0 {
		t.Fatalf("error/encrypt/store = %v/%d/%d", err, cipher.encryptCalls, store.calls)
	}
}

func TestProviderChannelServiceEncryptsValidatedContactBeforeStore(t *testing.T) {
	t.Parallel()
	owner := users.InternalUser{ID: uuid.New()}
	cipher := &providerCountingCipher{}
	store := &recordingChannelStore{}
	service := NewProviderChannelService(&recordingProviderAuthorizer{owner: owner}, store, cipher)
	status, err := service.Put(context.Background(), users.VerifiedIdentity{Subject: "provider"}, ReplaceChannel{Channel: ChannelWhatsApp, Contact: "+12025550123", Enabled: true, RevealConsent: true})
	if _, exists := reflect.TypeOf(store.value).FieldByName("Contact"); exists {
		t.Fatal("encrypted channel must not carry plaintext Contact")
	}
	if err != nil || status.Channel != ChannelWhatsApp || !status.Configured || cipher.encryptCalls != 1 || store.owner != owner.ID {
		t.Fatalf("status/error/cipher/store = %#v/%v/%d/%#v", status, err, cipher.encryptCalls, store)
	}
}

func TestProviderChannelServiceReadsStatusWithoutDecrypting(t *testing.T) {
	t.Parallel()
	owner := users.InternalUser{ID: uuid.New()}
	cipher := &providerCountingCipher{}
	store := &recordingChannelStore{statuses: []ChannelStatus{{Channel: ChannelPhone, Configured: true, Enabled: true, RevealConsent: true}}}
	service := NewProviderChannelService(&recordingProviderAuthorizer{owner: owner}, store, cipher)
	statuses, err := service.Get(context.Background(), users.VerifiedIdentity{Subject: "provider"})
	if err != nil || len(statuses) != 1 || statuses[0].Channel != ChannelPhone || cipher.decryptCalls != 0 || store.statusCalls != 1 {
		t.Fatalf("statuses/error/decrypt/store = %#v/%v/%d/%d", statuses, err, cipher.decryptCalls, store.statusCalls)
	}
}

type recordingProviderAuthorizer struct {
	owner users.InternalUser
	err   error
}

func (a *recordingProviderAuthorizer) RequireProvider(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return a.owner, a.err
}

type recordingChannelStore struct {
	calls       int
	owner       uuid.UUID
	value       EncryptedChannel
	statuses    []ChannelStatus
	statusCalls int
}

func (s *recordingChannelStore) Replace(_ context.Context, owner uuid.UUID, value EncryptedChannel) (ChannelStatus, error) {
	s.calls++
	s.owner = owner
	s.value = value
	return ChannelStatus{Channel: value.Channel, Configured: true, Enabled: value.Enabled, RevealConsent: value.RevealConsent}, nil
}

func (s *recordingChannelStore) Statuses(context.Context, uuid.UUID) ([]ChannelStatus, error) {
	s.statusCalls++
	return s.statuses, nil
}

type providerCountingCipher struct {
	decryptCalls int
	encryptCalls int
}

func (c *providerCountingCipher) Encrypt([]byte) (SealedContact, error) {
	c.encryptCalls++
	return SealedContact{Ciphertext: []byte{1}, Nonce: []byte{2}}, nil
}
func (c *providerCountingCipher) Decrypt(SealedContact) ([]byte, error) {
	c.decryptCalls++
	return []byte("revealed-contact"), nil
}
