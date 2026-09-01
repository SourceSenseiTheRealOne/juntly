package administration

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(db *sql.DB) Store { return sqlStore{database: db} }
func (s sqlStore) Metrics(ctx context.Context, actor uuid.UUID) (Metrics, error) {
	if ok, err := s.admin(ctx, s.database, actor); err != nil {
		return Metrics{}, err
	} else if !ok {
		return Metrics{}, ErrForbidden
	}
	var v Metrics
	err := s.database.QueryRowContext(ctx, `select (select count(*) from public.internal_users),(select count(*) from public.provider_profiles),(select count(*) from public.listings where state='active'),(select count(*) from public.bookings where state='completed'),(select count(*) from public.reviews where state='published'),(select count(*) from public.conversation_reports where state='open')`).Scan(&v.Users, &v.Providers, &v.ActiveListings, &v.CompletedBookings, &v.PublishedReviews, &v.OpenReports)
	return v, err
}
func (s sqlStore) Queue(ctx context.Context, actor uuid.UUID) (Queue, error) {
	if ok, err := s.admin(ctx, s.database, actor); err != nil {
		return Queue{}, err
	} else if !ok {
		return Queue{}, ErrForbidden
	}
	result := Queue{Reports: []ReportItem{}, Reviews: []ReviewItem{}}
	reports, err := s.database.QueryContext(ctx, `select id,conversation_id,reason,created_at from public.conversation_reports where state='open' order by created_at,id limit 100`)
	if err != nil {
		return Queue{}, err
	}
	defer reports.Close()
	for reports.Next() {
		var v ReportItem
		if err := reports.Scan(&v.ID, &v.ConversationID, &v.Reason, &v.CreatedAt); err != nil {
			return Queue{}, err
		}
		result.Reports = append(result.Reports, v)
	}
	if err = reports.Err(); err != nil {
		return Queue{}, err
	}
	reviews, err := s.database.QueryContext(ctx, `select id,rating,body,state,created_at from public.reviews order by created_at desc,id limit 100`)
	if err != nil {
		return Queue{}, err
	}
	defer reviews.Close()
	for reviews.Next() {
		var v ReviewItem
		if err := reviews.Scan(&v.ID, &v.Rating, &v.Body, &v.State, &v.CreatedAt); err != nil {
			return Queue{}, err
		}
		result.Reviews = append(result.Reviews, v)
	}
	return result, reviews.Err()
}
func (s sqlStore) Moderate(ctx context.Context, actor uuid.UUID, input ModerationAction) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if ok, err := s.admin(ctx, tx, actor); err != nil {
		return err
	} else if !ok {
		return ErrForbidden
	}
	targetType := "review"
	switch input.Kind {
	case "resolve_report":
		targetType = "conversation_report"
		result, err := tx.ExecContext(ctx, `update public.conversation_reports set state='resolved',resolved_at=timezone('utc',now()),resolved_by_internal_user_id=$1 where id=$2 and state='open'`, actor, input.TargetID)
		if err != nil {
			return err
		}
		if n, _ := result.RowsAffected(); n != 1 {
			return ErrNotFound
		}
	case "hide_review", "publish_review":
		var provider uuid.UUID
		var rating int
		var current string
		if err := tx.QueryRowContext(ctx, `select provider_internal_user_id,rating,state from public.reviews where id=$1 for update`, input.TargetID).Scan(&provider, &rating, &current); errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
		target := "hidden"
		delta := -1
		if input.Kind == "publish_review" {
			target = "published"
			delta = 1
		}
		if current == target {
			return ErrInvalid
		}
		if _, err := tx.ExecContext(ctx, `update public.reviews set state=$1,updated_at=timezone('utc',now()) where id=$2`, target, input.TargetID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `update public.provider_rating_aggregates set rating_sum=greatest(0,rating_sum+$1),review_count=greatest(0,review_count+$2),updated_at=timezone('utc',now()) where provider_internal_user_id=$3`, delta*rating, delta, provider); err != nil {
			return err
		}
	default:
		return ErrInvalid
	}
	if _, err := tx.ExecContext(ctx, `insert into public.administration_audit_records(actor_internal_user_id,action,target_type,target_id,reason) values($1,$2,$3,$4,$5)`, actor, input.Kind, targetType, input.TargetID, input.Reason); err != nil {
		return err
	}
	return tx.Commit()
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s sqlStore) admin(ctx context.Context, q queryer, actor uuid.UUID) (bool, error) {
	var ok bool
	err := q.QueryRowContext(ctx, `select exists(select 1 from public.platform_roles where internal_user_id=$1 and role='administrator')`, actor).Scan(&ok)
	return ok, err
}
