package reviews

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(db *sql.DB) Store { return sqlStore{database: db} }
func (s sqlStore) Create(ctx context.Context, actor uuid.UUID, input CreateReview) (Review, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Review{}, err
	}
	defer tx.Rollback()
	var provider uuid.UUID
	if err = tx.QueryRowContext(ctx, `select provider_internal_user_id from public.bookings where id=$1 and customer_internal_user_id=$2 and state='completed'`, input.BookingID, actor).Scan(&provider); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Review{}, ErrForbidden
		}
		return Review{}, err
	}
	var v Review
	err = tx.QueryRowContext(ctx, `insert into public.reviews(booking_id,customer_internal_user_id,provider_internal_user_id,rating,body) values($1,$2,$3,$4,$5) on conflict(booking_id) do nothing returning id,booking_id,customer_internal_user_id,provider_internal_user_id,rating,body,coalesce(provider_response,''),verified_booking,state,created_at,updated_at`, input.BookingID, actor, provider, input.Rating, input.Body).Scan(reviewScan(&v)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Review{}, ErrConflict
	}
	if err != nil {
		return Review{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.provider_rating_aggregates(provider_internal_user_id,rating_sum,review_count) values($1,$2,1) on conflict(provider_internal_user_id) do update set rating_sum=provider_rating_aggregates.rating_sum+excluded.rating_sum,review_count=provider_rating_aggregates.review_count+1,updated_at=timezone('utc',now())`, provider, input.Rating); err != nil {
		return Review{}, err
	}
	if err = s.notify(ctx, tx, provider, "review_received", v.ID); err != nil {
		return Review{}, err
	}
	if err = tx.Commit(); err != nil {
		return Review{}, err
	}
	return v, nil
}
func (s sqlStore) ListForProvider(ctx context.Context, actor uuid.UUID) ([]Review, error) {
	rows, err := s.database.QueryContext(ctx, `select id,booking_id,customer_internal_user_id,provider_internal_user_id,rating,body,coalesce(provider_response,''),verified_booking,state,created_at,updated_at from public.reviews where provider_internal_user_id=$1 order by created_at desc,id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Review{}
	for rows.Next() {
		var v Review
		if err := rows.Scan(reviewScan(&v)...); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (s sqlStore) Respond(ctx context.Context, actor, id uuid.UUID, response string) (Review, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Review{}, err
	}
	defer tx.Rollback()
	var v Review
	err = tx.QueryRowContext(ctx, `update public.reviews set provider_response=$1,updated_at=timezone('utc',now()) where id=$2 and provider_internal_user_id=$3 and provider_response is null returning id,booking_id,customer_internal_user_id,provider_internal_user_id,rating,body,provider_response,verified_booking,state,created_at,updated_at`, response, id, actor).Scan(reviewScan(&v)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Review{}, ErrForbidden
	}
	if err != nil {
		return Review{}, err
	}
	if err = s.notify(ctx, tx, v.CustomerID, "review_response", v.ID); err != nil {
		return Review{}, err
	}
	if err = tx.Commit(); err != nil {
		return Review{}, err
	}
	return v, nil
}
func (s sqlStore) Aggregate(ctx context.Context, provider uuid.UUID) (Aggregate, error) {
	var sum int64
	var count int
	err := s.database.QueryRowContext(ctx, `select coalesce(rating_sum,0),coalesce(review_count,0) from public.provider_rating_aggregates where provider_internal_user_id=$1`, provider).Scan(&sum, &count)
	if errors.Is(err, sql.ErrNoRows) {
		return Aggregate{ProviderID: provider}, nil
	}
	if err != nil {
		return Aggregate{}, err
	}
	average := 0.0
	if count > 0 {
		average = float64(sum) / float64(count)
	}
	return Aggregate{ProviderID: provider, AverageRating: average, ReviewCount: count}, nil
}
func (s sqlStore) notify(ctx context.Context, tx *sql.Tx, recipient uuid.UUID, kind string, resource uuid.UUID) error {
	if _, err := tx.ExecContext(ctx, `insert into public.notifications(recipient_internal_user_id,kind,resource_id,in_app_visible) select $1,$2,$3,coalesce(p.in_app_enabled,true) from (select 1)x left join public.notification_preferences p on p.internal_user_id=$1 where coalesce(p.in_app_enabled or p.email_enabled,true) on conflict do nothing`, recipient, kind, resource); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `insert into public.notification_email_outbox(notification_id,recipient_internal_user_id) select n.id,n.recipient_internal_user_id from public.notifications n left join public.notification_preferences p on p.internal_user_id=n.recipient_internal_user_id where n.recipient_internal_user_id=$1 and n.kind=$2 and n.resource_id=$3 and coalesce(p.email_enabled,true) on conflict(notification_id) do nothing`, recipient, kind, resource)
	return err
}
func reviewScan(v *Review) []any {
	return []any{&v.ID, &v.BookingID, &v.CustomerID, &v.ProviderID, &v.Rating, &v.Body, &v.ProviderResponse, &v.VerifiedBooking, &v.State, &v.CreatedAt, &v.UpdatedAt}
}
