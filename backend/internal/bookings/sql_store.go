package bookings

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(database *sql.DB) Store { return sqlStore{database: database} }
func (s sqlStore) Create(ctx context.Context, actor uuid.UUID, input CreateBooking) (Booking, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback()
	provider, price, err := s.resolveSource(ctx, tx, actor, input)
	if err != nil {
		return Booking{}, err
	}
	var v Booking
	err = tx.QueryRowContext(ctx, `insert into public.bookings(customer_internal_user_id,provider_internal_user_id,source_type,source_id,idempotency_key,scheduled_at,private_location,agreed_price_minor) values($1,$2,$3,$4,$5,$6,$7,$8) on conflict do nothing returning id,customer_internal_user_id,provider_internal_user_id,source_type,source_id,state,revision,scheduled_at,private_location,agreed_price_minor,currency,created_at,updated_at`, actor, provider, input.SourceType, input.SourceID, input.IdempotencyKey, input.ScheduledAt, input.PrivateLocation, price).Scan(bookingScan(&v)...)
	if errors.Is(err, sql.ErrNoRows) {
		existing, e := s.getByIdempotency(ctx, tx, actor, input.IdempotencyKey)
		if e != nil {
			if errors.Is(e, sql.ErrNoRows) && input.SourceType == SourceProposal {
				return Booking{}, ErrConflict
			}
			return Booking{}, e
		}
		if existing.SourceType != input.SourceType || !sameUUID(existing.SourceID, input.SourceID) || existing.ProviderID != provider {
			return Booking{}, ErrConflict
		}
		return existing, tx.Commit()
	}
	if err != nil {
		return Booking{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.booking_events(booking_id,actor_internal_user_id,from_state,to_state,revision) values($1,$2,null,$3,1)`, v.ID, actor, v.State); err != nil {
		return Booking{}, err
	}
	if err = s.notify(ctx, tx, v.ProviderID, "booking_created", v.ID); err != nil {
		return Booking{}, err
	}
	if err = tx.Commit(); err != nil {
		return Booking{}, err
	}
	return v, nil
}
func (s sqlStore) resolveSource(ctx context.Context, tx *sql.Tx, actor uuid.UUID, input CreateBooking) (uuid.UUID, int, error) {
	switch input.SourceType {
	case SourceProposal:
		var provider uuid.UUID
		var price int
		err := tx.QueryRowContext(ctx, `select p.provider_internal_user_id,p.price_minor from public.quotation_proposals p join public.quotation_requests r on r.id=p.request_id where p.id=$1 and p.state='accepted' and r.customer_internal_user_id=$2`, *input.SourceID, actor).Scan(&provider, &price)
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, 0, ErrForbidden
		}
		return provider, price, err
	case SourceListing:
		var provider uuid.UUID
		var listedPrice *int
		var priceType string
		err := tx.QueryRowContext(ctx, `select internal_user_id,price_minor,price_type from public.listings where id=$1 and state='active' and internal_user_id<>$2`, *input.SourceID, actor).Scan(&provider, &listedPrice, &priceType)
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, 0, ErrNotFound
		}
		if err != nil {
			return uuid.Nil, 0, err
		}
		if listedPrice != nil {
			return provider, *listedPrice, nil
		}
		if input.AgreedPriceMinor == nil || *input.AgreedPriceMinor <= 0 {
			return uuid.Nil, 0, ErrInvalid
		}
		return provider, *input.AgreedPriceMinor, nil
	case SourceDirect:
		var active bool
		err := tx.QueryRowContext(ctx, `select exists(select 1 from public.provider_profiles where internal_user_id=$1)`, *input.ProviderID).Scan(&active)
		if err != nil {
			return uuid.Nil, 0, err
		}
		if !active || *input.ProviderID == actor {
			return uuid.Nil, 0, ErrForbidden
		}
		return *input.ProviderID, *input.AgreedPriceMinor, nil
	default:
		return uuid.Nil, 0, ErrInvalid
	}
}
func (s sqlStore) List(ctx context.Context, actor uuid.UUID) ([]Booking, error) {
	rows, err := s.database.QueryContext(ctx, `select id,customer_internal_user_id,provider_internal_user_id,source_type,source_id,state,revision,scheduled_at,private_location,agreed_price_minor,currency,created_at,updated_at from public.bookings where customer_internal_user_id=$1 or provider_internal_user_id=$1 order by updated_at desc,id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Booking{}
	for rows.Next() {
		var v Booking
		if err := rows.Scan(bookingScan(&v)...); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (s sqlStore) Get(ctx context.Context, actor, id uuid.UUID) (Booking, error) {
	var v Booking
	err := s.database.QueryRowContext(ctx, `select id,customer_internal_user_id,provider_internal_user_id,source_type,source_id,state,revision,scheduled_at,private_location,agreed_price_minor,currency,created_at,updated_at from public.bookings where id=$1 and (customer_internal_user_id=$2 or provider_internal_user_id=$2)`, id, actor).Scan(bookingScan(&v)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Booking{}, ErrForbidden
	}
	return v, err
}
func (s sqlStore) Transition(ctx context.Context, actor, id uuid.UUID, input Transition) (Booking, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback()
	var current Booking
	err = tx.QueryRowContext(ctx, `select id,customer_internal_user_id,provider_internal_user_id,source_type,source_id,state,revision,scheduled_at,private_location,agreed_price_minor,currency,created_at,updated_at from public.bookings where id=$1 for update`, id).Scan(bookingScan(&current)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Booking{}, ErrNotFound
	}
	if err != nil {
		return Booking{}, err
	}
	if actor != current.CustomerID && actor != current.ProviderID && !s.moderator(ctx, tx, actor) {
		return Booking{}, ErrForbidden
	}
	if current.State == input.TargetState && current.Revision == input.Revision+1 {
		return current, tx.Commit()
	}
	if current.State != input.ExpectedState || current.Revision != input.Revision {
		return Booking{}, ErrConflict
	}
	if !authorizedTransition(ctx, tx, actor, current, input.TargetState) {
		return Booking{}, ErrForbidden
	}
	result, err := tx.ExecContext(ctx, `update public.bookings set state=$1,revision=revision+1,updated_at=timezone('utc',now()) where id=$2 and state=$3 and revision=$4`, input.TargetState, id, input.ExpectedState, input.Revision)
	if err != nil {
		return Booking{}, err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return Booking{}, ErrConflict
	}
	if _, err = tx.ExecContext(ctx, `insert into public.booking_events(booking_id,actor_internal_user_id,from_state,to_state,revision,reason) values($1,$2,$3,$4,$5,$6)`, id, actor, input.ExpectedState, input.TargetState, input.Revision+1, input.Reason); err != nil {
		return Booking{}, err
	}
	recipient := current.CustomerID
	if actor == current.CustomerID {
		recipient = current.ProviderID
	}
	if err = s.notify(ctx, tx, recipient, "booking_updated", id); err != nil {
		return Booking{}, err
	}
	current.State = input.TargetState
	current.Revision++
	if err = tx.Commit(); err != nil {
		return Booking{}, err
	}
	return current, nil
}
func authorizedTransition(ctx context.Context, tx *sql.Tx, actor uuid.UUID, b Booking, target State) bool {
	switch target {
	case StateConfirmed, StateInProgress, StateCompleted:
		return actor == b.ProviderID
	case StateRefunded:
		var ok bool
		_ = tx.QueryRowContext(ctx, `select exists(select 1 from public.platform_roles where internal_user_id=$1 and role='moderator')`, actor).Scan(&ok)
		return ok
	default:
		return actor == b.CustomerID || actor == b.ProviderID
	}
}
func (s sqlStore) moderator(ctx context.Context, tx *sql.Tx, actor uuid.UUID) bool {
	var ok bool
	_ = tx.QueryRowContext(ctx, `select exists(select 1 from public.platform_roles where internal_user_id=$1 and role='moderator')`, actor).Scan(&ok)
	return ok
}
func (s sqlStore) notify(ctx context.Context, tx *sql.Tx, recipient uuid.UUID, kind string, resource uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `insert into public.notifications(recipient_internal_user_id,kind,resource_id,in_app_visible) select $1,$2,$3,coalesce(p.in_app_enabled,true) from (select 1) x left join public.notification_preferences p on p.internal_user_id=$1 where coalesce(p.in_app_enabled or p.email_enabled,true) on conflict do nothing`, recipient, kind, resource); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `insert into public.notification_email_outbox(notification_id,recipient_internal_user_id) select n.id,n.recipient_internal_user_id from public.notifications n left join public.notification_preferences p on p.internal_user_id=n.recipient_internal_user_id where n.recipient_internal_user_id=$1 and n.kind=$2 and n.resource_id=$3 and coalesce(p.email_enabled,true) on conflict(notification_id) do nothing`, recipient, kind, resource)
	return err
}
func (s sqlStore) getByIdempotency(ctx context.Context, tx *sql.Tx, actor uuid.UUID, key string) (Booking, error) {
	var v Booking
	err := tx.QueryRowContext(ctx, `select id,customer_internal_user_id,provider_internal_user_id,source_type,source_id,state,revision,scheduled_at,private_location,agreed_price_minor,currency,created_at,updated_at from public.bookings where customer_internal_user_id=$1 and idempotency_key=$2`, actor, key).Scan(bookingScan(&v)...)
	return v, err
}
func bookingScan(v *Booking) []any {
	return []any{&v.ID, &v.CustomerID, &v.ProviderID, &v.SourceType, &v.SourceID, &v.State, &v.Revision, &v.ScheduledAt, &v.PrivateLocation, &v.AgreedPriceMinor, &v.Currency, &v.CreatedAt, &v.UpdatedAt}
}
func sameUUID(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
