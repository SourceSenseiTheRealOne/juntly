package contactreveal

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type sqlChannelStore struct{ database *sql.DB }

func NewSQLChannelStore(database *sql.DB) ChannelStore {
	return sqlChannelStore{database: database}
}

func (s sqlChannelStore) Replace(ctx context.Context, ownerID uuid.UUID, value EncryptedChannel) (ChannelStatus, error) {
	if s.database == nil || ownerID == uuid.Nil || !validChannel(value.Channel) || len(value.Sealed.Ciphertext) == 0 || len(value.Sealed.Nonce) == 0 || value.KeyVersion == "" {
		return ChannelStatus{}, ErrUnavailable
	}
	_, err := s.database.ExecContext(ctx, `
		insert into public.provider_contact_channels
		  (id, internal_user_id, channel, ciphertext, nonce, key_version, enabled, reveal_consent)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (internal_user_id, channel) do update set
		  ciphertext = excluded.ciphertext,
		  nonce = excluded.nonce,
		  key_version = excluded.key_version,
		  enabled = excluded.enabled,
		  reveal_consent = excluded.reveal_consent,
		  updated_at = timezone('utc', now())
	`, uuid.New(), ownerID, string(value.Channel), value.Sealed.Ciphertext, value.Sealed.Nonce, value.KeyVersion, value.Enabled, value.RevealConsent)
	if err != nil {
		return ChannelStatus{}, ErrUnavailable
	}
	return ChannelStatus{Channel: value.Channel, Configured: true, Enabled: value.Enabled, RevealConsent: value.RevealConsent}, nil
}
