package contactreveal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestRevealServiceDeniesBeforeDecryptingContact(t *testing.T) {
	t.Parallel()
	identity := users.VerifiedIdentity{Subject: "customer"}
	store := &recordingRevealStore{err: ErrForbidden}
	cipher := &countingCipher{}
	service := NewRevealService(&recordingIdentityReconciler{user: users.InternalUser{ID: uuid.New()}}, store, cipher, func() time.Time { return time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC) })
	_, err := service.Reveal(context.Background(), identity, uuid.New(), ChannelPhone)
	if !errors.Is(err, ErrForbidden) || cipher.decryptCalls != 0 || store.calls != 1 {
		t.Fatalf("error/decrypt/store = %v/%d/%d", err, cipher.decryptCalls, store.calls)
	}
}

func TestRevealServiceDecryptsOnlyAuthorizedReservedContact(t *testing.T) {
	t.Parallel()
	day := time.Date(2026, 8, 24, 17, 30, 0, 0, time.FixedZone("offset", 3600))
	store := &recordingRevealStore{sealed: SealedContact{Ciphertext: []byte{1}, Nonce: []byte{2}}}
	cipher := &countingCipher{}
	service := NewRevealService(&recordingIdentityReconciler{user: users.InternalUser{ID: uuid.New()}}, store, cipher, func() time.Time { return day })
	value, err := service.Reveal(context.Background(), users.VerifiedIdentity{Subject: "customer"}, uuid.New(), ChannelWhatsApp)
	wantDay := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	if err != nil || value.Channel != ChannelWhatsApp || value.Value == "" || cipher.decryptCalls != 1 || !store.day.Equal(wantDay) {
		t.Fatalf("value/error/decrypt/day = %#v/%v/%d/%s", value, err, cipher.decryptCalls, store.day)
	}
}

type recordingIdentityReconciler struct{ user users.InternalUser }

func (r *recordingIdentityReconciler) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	return r.user, false, nil
}

type recordingRevealStore struct {
	calls  int
	err    error
	sealed SealedContact
	day    time.Time
}

func (r *recordingRevealStore) AuthorizeAndReserve(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ Channel, day time.Time) (SealedContact, error) {
	r.calls++
	r.day = day
	return r.sealed, r.err
}

type countingCipher struct{ decryptCalls int }

func (*countingCipher) Encrypt([]byte) (SealedContact, error) { return SealedContact{}, nil }
func (c *countingCipher) Decrypt(SealedContact) ([]byte, error) {
	c.decryptCalls++
	return []byte("revealed-contact"), nil
}
