package contactreveal

import (
	"context"
	"database/sql"
	"encoding/base64"
	"testing"

	"github.com/google/uuid"
)

func TestSQLChannelStoreUpsertsEncryptedOwnerChannelWithoutPlaintext(t *testing.T) {
	database := openRevealDatabase(t)
	ctx := context.Background()
	_, listingID, _ := seedRevealFixture(t, database)
	var providerID uuid.UUID
	if err := database.QueryRowContext(ctx, `select internal_user_id from public.listings where id = $1`, listingID).Scan(&providerID); err != nil {
		t.Fatalf("provider: %v", err)
	}
	key := make([]byte, 32)
	for index := range key {
		key[index] = byte(index + 1)
	}
	cipher, err := NewCipher(base64.StdEncoding.EncodeToString(key))
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	sealed, err := cipher.Encrypt([]byte("test-contact"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	status, err := NewSQLChannelStore(database).Replace(ctx, providerID, EncryptedChannel{Channel: ChannelPhone, Sealed: sealed, KeyVersion: "v1", Enabled: true, RevealConsent: true})
	if err != nil || !status.Configured || !status.Enabled || !status.RevealConsent || status.Channel != ChannelPhone {
		t.Fatalf("status/error = %#v/%v", status, err)
	}
	var ciphertext, nonce []byte
	if err := database.QueryRowContext(ctx, `select ciphertext, nonce from public.provider_contact_channels where internal_user_id = $1 and channel = 'phone'`, providerID).Scan(&ciphertext, &nonce); err != nil {
		t.Fatalf("stored channel: %v", err)
	}
	if string(ciphertext) != string(sealed.Ciphertext) || string(nonce) != string(sealed.Nonce) {
		t.Fatal("stored ciphertext or nonce mismatch")
	}
}

var _ = sql.ErrNoRows
