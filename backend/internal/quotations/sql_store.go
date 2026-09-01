package quotations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(database *sql.DB) Store { return sqlStore{database: database} }
func (s sqlStore) CreateRequest(ctx context.Context, actor uuid.UUID, input CreateRequest) (Request, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Request{}, err
	}
	defer tx.Rollback()
	var v Request
	err = tx.QueryRowContext(ctx, `insert into public.quotation_requests (customer_internal_user_id,category_id,locality_id,title,description,budget_minor,proposal_deadline) select $1,$2,$3,$4,$5,$6,$7 where exists(select 1 from public.service_categories where id=$2 and active) and exists(select 1 from public.localities where id=$3 and active) returning id,customer_internal_user_id,title,description,category_id,locality_id,budget_minor,proposal_deadline,state,created_at,updated_at`, actor, input.CategoryID, input.LocalityID, input.Title, input.Description, input.BudgetMinor, input.ProposalDeadline).Scan(&v.ID, &v.CustomerID, &v.Title, &v.Description, &v.CategoryID, &v.LocalityID, &v.BudgetMinor, &v.ProposalDeadline, &v.State, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Request{}, ErrInvalid
	}
	if err != nil {
		return Request{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notifications(recipient_internal_user_id,kind,resource_id,in_app_visible)
		select distinct l.internal_user_id,'request_published',$1,coalesce(p.in_app_enabled,true)
		from public.listings l left join public.provider_service_localities psl on psl.internal_user_id=l.internal_user_id and psl.locality_id=$3
		left join public.notification_preferences p on p.internal_user_id=l.internal_user_id
		where l.category_id=$2 and l.state='active' and l.internal_user_id<>$4 and (l.primary_locality_id=$3 or psl.locality_id is not null)
		and coalesce(p.in_app_enabled or p.email_enabled,true) on conflict do nothing`, v.ID, v.CategoryID, v.LocalityID, actor); err != nil {
		return Request{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notification_email_outbox(notification_id,recipient_internal_user_id)
		select n.id,n.recipient_internal_user_id from public.notifications n left join public.notification_preferences p on p.internal_user_id=n.recipient_internal_user_id
		where n.kind='request_published' and n.resource_id=$1 and coalesce(p.email_enabled,true) on conflict(notification_id) do nothing`, v.ID); err != nil {
		return Request{}, err
	}
	if err = tx.Commit(); err != nil {
		return Request{}, err
	}
	return v, nil
}
func (s sqlStore) ListCustomerRequests(ctx context.Context, actor uuid.UUID) ([]Request, error) {
	rows, err := s.database.QueryContext(ctx, `select id,customer_internal_user_id,title,description,category_id,locality_id,budget_minor,proposal_deadline,state,created_at,updated_at from public.quotation_requests where customer_internal_user_id=$1 order by updated_at desc,id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}
func (s sqlStore) ListOpportunities(ctx context.Context, actor uuid.UUID) ([]Request, error) {
	rows, err := s.database.QueryContext(ctx, `select distinct r.id,r.customer_internal_user_id,r.title,r.description,r.category_id,r.locality_id,r.budget_minor,r.proposal_deadline,r.state,r.created_at,r.updated_at from public.quotation_requests r join public.listings l on l.internal_user_id=$1 and l.category_id=r.category_id and l.state='active' left join public.provider_service_localities psl on psl.internal_user_id=$1 and psl.locality_id=r.locality_id where r.state='open' and r.proposal_deadline>timezone('utc',now()) and r.customer_internal_user_id<>$1 and (l.primary_locality_id=r.locality_id or psl.locality_id is not null) order by r.updated_at desc,r.id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRequests(rows)
}
func (s sqlStore) SubmitProposal(ctx context.Context, actor, requestID uuid.UUID, input SubmitProposal) (Proposal, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Proposal{}, err
	}
	defer tx.Rollback()
	var v Proposal
	err = tx.QueryRowContext(ctx, `insert into public.quotation_proposals (request_id,provider_internal_user_id,price_minor,message,available_at,estimated_minutes,expires_at) select r.id,$1,$3,$4,$5,$6,$7 from public.quotation_requests r where r.id=$2 and r.state='open' and r.proposal_deadline>timezone('utc',now()) and r.customer_internal_user_id<>$1 and exists(select 1 from public.listings l left join public.provider_service_localities psl on psl.internal_user_id=$1 and psl.locality_id=r.locality_id where l.internal_user_id=$1 and l.category_id=r.category_id and l.state='active' and (l.primary_locality_id=r.locality_id or psl.locality_id is not null)) on conflict(request_id,provider_internal_user_id) do nothing returning id,request_id,provider_internal_user_id,price_minor,message,available_at,estimated_minutes,expires_at,state,created_at,updated_at`, actor, requestID, input.PriceMinor, input.Message, input.AvailableAt, input.EstimatedMinutes, input.ExpiresAt).Scan(&v.ID, &v.RequestID, &v.ProviderID, &v.PriceMinor, &v.Message, &v.AvailableAt, &v.EstimatedMinutes, &v.ExpiresAt, &v.State, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Proposal{}, ErrForbidden
	}
	if err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notifications(recipient_internal_user_id,kind,resource_id,in_app_visible)
		select r.customer_internal_user_id,'proposal_received',$1,coalesce(p.in_app_enabled,true) from public.quotation_requests r
		left join public.notification_preferences p on p.internal_user_id=r.customer_internal_user_id where r.id=$2
		and coalesce(p.in_app_enabled or p.email_enabled,true) on conflict do nothing`, v.ID, requestID); err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notification_email_outbox(notification_id,recipient_internal_user_id)
		select n.id,n.recipient_internal_user_id from public.notifications n left join public.notification_preferences p on p.internal_user_id=n.recipient_internal_user_id
		where n.kind='proposal_received' and n.resource_id=$1 and coalesce(p.email_enabled,true) on conflict(notification_id) do nothing`, v.ID); err != nil {
		return Proposal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Proposal{}, err
	}
	return v, nil
}
func (s sqlStore) ListProposals(ctx context.Context, actor, requestID uuid.UUID) ([]Proposal, error) {
	if _, err := s.database.ExecContext(ctx, `update public.quotation_proposals set state='expired',updated_at=timezone('utc',now()) where request_id=$1 and state='submitted' and expires_at is not null and expires_at<=timezone('utc',now())`, requestID); err != nil {
		return nil, err
	}
	rows, err := s.database.QueryContext(ctx, `select p.id,p.request_id,p.provider_internal_user_id,p.price_minor,p.message,p.available_at,p.estimated_minutes,p.expires_at,p.state,p.created_at,p.updated_at from public.quotation_proposals p join public.quotation_requests r on r.id=p.request_id where p.request_id=$1 and (r.customer_internal_user_id=$2 or p.provider_internal_user_id=$2) order by p.created_at,p.id`, requestID, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values, err := scanProposals(rows)
	if err == nil && len(values) == 0 {
		var allowed bool
		if e := s.database.QueryRowContext(ctx, `select exists(select 1 from public.quotation_requests where id=$1 and customer_internal_user_id=$2)`, requestID, actor).Scan(&allowed); e != nil {
			return nil, e
		}
		if !allowed {
			return nil, ErrForbidden
		}
	}
	return values, err
}
func (s sqlStore) AcceptProposal(ctx context.Context, actor, requestID, proposalID uuid.UUID) (Proposal, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Proposal{}, err
	}
	defer tx.Rollback()
	var state RequestState
	if err = tx.QueryRowContext(ctx, `select state from public.quotation_requests where id=$1 and customer_internal_user_id=$2 for update`, requestID, actor).Scan(&state); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Proposal{}, ErrForbidden
		}
		return Proposal{}, err
	}
	if state != RequestOpen {
		return Proposal{}, ErrConflict
	}
	var v Proposal
	err = tx.QueryRowContext(ctx, `update public.quotation_proposals set state='accepted',updated_at=timezone('utc',now()) where id=$1 and request_id=$2 and state='submitted' and (expires_at is null or expires_at>timezone('utc',now())) returning id,request_id,provider_internal_user_id,price_minor,message,available_at,estimated_minutes,expires_at,state,created_at,updated_at`, proposalID, requestID).Scan(&v.ID, &v.RequestID, &v.ProviderID, &v.PriceMinor, &v.Message, &v.AvailableAt, &v.EstimatedMinutes, &v.ExpiresAt, &v.State, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Proposal{}, ErrConflict
	}
	if err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `update public.quotation_proposals set state='rejected',updated_at=timezone('utc',now()) where request_id=$1 and id<>$2 and state='submitted'`, requestID, proposalID); err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `update public.quotation_requests set state='accepted',updated_at=timezone('utc',now()) where id=$1`, requestID); err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notifications(recipient_internal_user_id,kind,resource_id,in_app_visible)
		select qp.provider_internal_user_id,case when qp.id=$2 then 'proposal_accepted' else 'proposal_rejected' end,qp.id,coalesce(np.in_app_enabled,true)
		from public.quotation_proposals qp left join public.notification_preferences np on np.internal_user_id=qp.provider_internal_user_id
		where qp.request_id=$1 and qp.state in ('accepted','rejected') and coalesce(np.in_app_enabled or np.email_enabled,true) on conflict do nothing`, requestID, proposalID); err != nil {
		return Proposal{}, err
	}
	if _, err = tx.ExecContext(ctx, `insert into public.notification_email_outbox(notification_id,recipient_internal_user_id)
		select n.id,n.recipient_internal_user_id from public.notifications n left join public.notification_preferences p on p.internal_user_id=n.recipient_internal_user_id
		where n.kind in ('proposal_accepted','proposal_rejected') and n.resource_id in (select id from public.quotation_proposals where request_id=$1)
		and coalesce(p.email_enabled,true) on conflict(notification_id) do nothing`, requestID); err != nil {
		return Proposal{}, err
	}
	if err = tx.Commit(); err != nil {
		return Proposal{}, err
	}
	return v, nil
}
func scanRequests(rows *sql.Rows) ([]Request, error) {
	values := []Request{}
	for rows.Next() {
		var v Request
		if err := rows.Scan(&v.ID, &v.CustomerID, &v.Title, &v.Description, &v.CategoryID, &v.LocalityID, &v.BudgetMinor, &v.ProposalDeadline, &v.State, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func scanProposals(rows *sql.Rows) ([]Proposal, error) {
	values := []Proposal{}
	for rows.Next() {
		var v Proposal
		if err := rows.Scan(&v.ID, &v.RequestID, &v.ProviderID, &v.PriceMinor, &v.Message, &v.AvailableAt, &v.EstimatedMinutes, &v.ExpiresAt, &v.State, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
