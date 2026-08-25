package contactreveal

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type sqlRevealStore struct{ database *sql.DB }

func NewSQLRevealStore(database *sql.DB) RevealStore {
	return sqlRevealStore{database: database}
}

func (s sqlRevealStore) AuthorizeAndReserve(ctx context.Context, customerID, listingID uuid.UUID, channel Channel, day time.Time) (SealedContact, error) {
	if s.database == nil || customerID == uuid.Nil || listingID == uuid.Nil || !validChannel(channel) {
		return SealedContact{}, ErrForbidden
	}
	transaction, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return SealedContact{}, ErrUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()
	var providerID uuid.UUID
	if err := transaction.QueryRowContext(ctx, `
		select internal_user_id from public.listings
		where id = $1 and state = 'active'
		for key share
	`, listingID).Scan(&providerID); err != nil {
		return SealedContact{}, policyError(err)
	}
	if providerID == customerID {
		return SealedContact{}, ErrForbidden
	}
	var sealed SealedContact
	if err := transaction.QueryRowContext(ctx, `
		select ciphertext, nonce from public.provider_contact_channels
		where internal_user_id = $1 and channel = $2 and enabled and reveal_consent
		for key share
	`, providerID, string(channel)).Scan(&sealed.Ciphertext, &sealed.Nonce); err != nil {
		return SealedContact{}, policyError(err)
	}
	if _, err := transaction.ExecContext(ctx, `
		insert into public.contact_reveal_daily_limits (id, customer_internal_user_id, utc_day, successful_count)
		values ($1, $2, $3, 0)
		on conflict (customer_internal_user_id, utc_day) do nothing
	`, uuid.New(), customerID, day); err != nil {
		return SealedContact{}, ErrUnavailable
	}
	var count int
	if err := transaction.QueryRowContext(ctx, `
		select successful_count from public.contact_reveal_daily_limits
		where customer_internal_user_id = $1 and utc_day = $2
		for update
	`, customerID, day).Scan(&count); err != nil {
		return SealedContact{}, ErrUnavailable
	}
	var existing bool
	err = transaction.QueryRowContext(ctx, `
		select true from public.contact_reveal_events
		where customer_internal_user_id = $1 and listing_id = $2 and channel = $3 and utc_day = $4
		limit 1
	`, customerID, listingID, string(channel), day).Scan(&existing)
	if err == nil && existing {
		if err := transaction.Commit(); err != nil {
			return SealedContact{}, ErrUnavailable
		}
		committed = true
		return copySealed(sealed), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return SealedContact{}, ErrUnavailable
	}
	if count >= 10 {
		return SealedContact{}, ErrForbidden
	}
	if _, err := transaction.ExecContext(ctx, `
		update public.contact_reveal_daily_limits
		set successful_count = successful_count + 1, updated_at = timezone('utc', now())
		where customer_internal_user_id = $1 and utc_day = $2 and successful_count < 10
	`, customerID, day); err != nil {
		return SealedContact{}, ErrUnavailable
	}
	if _, err := transaction.ExecContext(ctx, `
		insert into public.contact_reveal_events
		  (id, customer_internal_user_id, provider_internal_user_id, listing_id, channel, utc_day)
		values ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), customerID, providerID, listingID, string(channel), day); err != nil {
		return SealedContact{}, ErrUnavailable
	}
	if err := transaction.Commit(); err != nil {
		return SealedContact{}, ErrUnavailable
	}
	committed = true
	return copySealed(sealed), nil
}

func policyError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	return ErrUnavailable
}

func copySealed(value SealedContact) SealedContact {
	return SealedContact{Ciphertext: append([]byte(nil), value.Ciphertext...), Nonce: append([]byte(nil), value.Nonce...)}
}
