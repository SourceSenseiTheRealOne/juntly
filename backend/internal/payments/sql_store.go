package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(database *sql.DB) Store { return sqlStore{database: database} }

func (s sqlStore) PrepareCheckout(ctx context.Context, actor, bookingID uuid.UUID, key string, feeBPS int) (Order, ProviderAccount, error) {
	if s.database == nil {
		return Order{}, ProviderAccount{}, ErrUnavailable
	}
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, ProviderAccount{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var customer, provider uuid.UUID
	var bookingState string
	var gross int64
	var currency string
	var account ProviderAccount
	err = tx.QueryRowContext(ctx, `select b.customer_internal_user_id,b.provider_internal_user_id,b.state,b.agreed_price_minor,b.currency,coalesce(a.stripe_account_id,''),coalesce(a.details_submitted,false),coalesce(a.charges_enabled,false),coalesce(a.payouts_enabled,false),coalesce(a.updated_at,timezone('utc',now())) from public.bookings b left join public.provider_payment_accounts a on a.internal_user_id=b.provider_internal_user_id where b.id=$1 for update of b`, bookingID).Scan(&customer, &provider, &bookingState, &gross, &currency, &account.StripeAccountID, &account.DetailsSubmitted, &account.ChargesEnabled, &account.PayoutsEnabled, &account.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, ProviderAccount{}, ErrNotFound
	}
	if err != nil {
		return Order{}, ProviderAccount{}, err
	}
	if customer != actor {
		return Order{}, ProviderAccount{}, ErrForbidden
	}
	if bookingState != "confirmed" && bookingState != "scheduled" {
		return Order{}, ProviderAccount{}, ErrConflict
	}
	account.InternalUserID = provider.String()
	if account.StripeAccountID == "" {
		return Order{}, ProviderAccount{}, ErrForbidden
	}
	if existing, found, loadErr := loadOrderByBooking(ctx, tx, bookingID); loadErr != nil {
		return Order{}, ProviderAccount{}, loadErr
	} else if found {
		var existingKey string
		if err := tx.QueryRowContext(ctx, `select idempotency_key from public.payment_orders where id=$1`, existing.ID).Scan(&existingKey); err != nil {
			return Order{}, ProviderAccount{}, err
		}
		if existingKey != key {
			return Order{}, ProviderAccount{}, ErrConflict
		}
		if err := tx.Commit(); err != nil {
			return Order{}, ProviderAccount{}, err
		}
		return existing, account, nil
	}
	fee := gross * int64(feeBPS) / 10_000
	if fee < 0 || fee >= gross || currency != "EUR" {
		return Order{}, ProviderAccount{}, ErrInvalid
	}
	id := uuid.New()
	var order Order
	err = tx.QueryRowContext(ctx, `insert into public.payment_orders(id,booking_id,customer_internal_user_id,provider_internal_user_id,idempotency_key,gross_minor,platform_fee_minor,provider_net_minor,currency) values($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id,booking_id,customer_internal_user_id,provider_internal_user_id,state,gross_minor,platform_fee_minor,provider_net_minor,currency,coalesce(stripe_checkout_session_id,''),coalesce(stripe_payment_intent_id,''),coalesce(stripe_invoice_id,''),coalesce(stripe_refund_id,''),created_at,updated_at`, id, bookingID, customer, provider, key, gross, fee, gross-fee, currency).Scan(orderScan(&order)...)
	if err != nil {
		return Order{}, ProviderAccount{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, ProviderAccount{}, err
	}
	return order, account, nil
}

func (s sqlStore) AttachCheckout(ctx context.Context, actor, orderID uuid.UUID, session CheckoutSession) (Order, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	order, found, err := loadOrderByID(ctx, tx, orderID, true)
	if err != nil || !found {
		if !found && err == nil {
			err = ErrNotFound
		}
		return Order{}, err
	}
	if order.CustomerID != actor.String() {
		return Order{}, ErrForbidden
	}
	if order.State == StateCheckoutCreated && order.CheckoutSessionID == session.ID {
		if err := tx.Commit(); err != nil {
			return Order{}, err
		}
		return order, nil
	}
	if order.State != StatePendingCheckout {
		return Order{}, ErrConflict
	}
	result, err := tx.ExecContext(ctx, `update public.payment_orders set state='checkout_created',stripe_checkout_session_id=$1,updated_at=timezone('utc',now()) where id=$2 and state='pending_checkout'`, session.ID, orderID)
	if err != nil {
		return Order{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Order{}, ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `insert into public.payment_events(payment_order_id,event_type,from_state,to_state,provider_object_id) values($1,'checkout_created','pending_checkout','checkout_created',$2)`, orderID, session.ID); err != nil {
		return Order{}, err
	}
	order, _, err = loadOrderByID(ctx, tx, orderID, false)
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return order, nil
}

func (s sqlStore) ListOrders(ctx context.Context, actor uuid.UUID) ([]Order, error) {
	rows, err := s.database.QueryContext(ctx, `select id,booking_id,customer_internal_user_id,provider_internal_user_id,state,gross_minor,platform_fee_minor,provider_net_minor,currency,coalesce(stripe_checkout_session_id,''),coalesce(stripe_payment_intent_id,''),coalesce(stripe_invoice_id,''),coalesce(stripe_refund_id,''),created_at,updated_at from public.payment_orders where customer_internal_user_id=$1 or provider_internal_user_id=$1 order by updated_at desc,id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var order Order
		if err := rows.Scan(orderScan(&order)...); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (s sqlStore) ListAdminOrders(ctx context.Context, actor uuid.UUID) ([]Order, error) {
	var authorized bool
	if err := s.database.QueryRowContext(ctx, `select exists(select 1 from public.platform_roles where internal_user_id=$1 and role in('moderator','administrator'))`, actor).Scan(&authorized); err != nil {
		return nil, err
	}
	if !authorized {
		return nil, ErrForbidden
	}
	rows, err := s.database.QueryContext(ctx, orderSelect+` order by updated_at desc,id limit 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var order Order
		if err := rows.Scan(orderScan(&order)...); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (s sqlStore) GetProviderAccount(ctx context.Context, actor uuid.UUID) (ProviderAccount, error) {
	var account ProviderAccount
	err := s.database.QueryRowContext(ctx, `select internal_user_id,stripe_account_id,details_submitted,charges_enabled,payouts_enabled,updated_at from public.provider_payment_accounts where internal_user_id=$1`, actor).Scan(&account.InternalUserID, &account.StripeAccountID, &account.DetailsSubmitted, &account.ChargesEnabled, &account.PayoutsEnabled, &account.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ProviderAccount{}, ErrNotFound
	}
	return account, err
}

func (s sqlStore) SaveProviderAccount(ctx context.Context, actor uuid.UUID, value ConnectedAccount) (ProviderAccount, error) {
	var providerExists bool
	if err := s.database.QueryRowContext(ctx, `select exists(select 1 from public.provider_profiles where internal_user_id=$1)`, actor).Scan(&providerExists); err != nil {
		return ProviderAccount{}, err
	}
	if !providerExists {
		return ProviderAccount{}, ErrForbidden
	}
	var account ProviderAccount
	err := s.database.QueryRowContext(ctx, `insert into public.provider_payment_accounts(internal_user_id,stripe_account_id,details_submitted,charges_enabled,payouts_enabled) values($1,$2,$3,$4,$5) on conflict(internal_user_id) do update set details_submitted=excluded.details_submitted,charges_enabled=excluded.charges_enabled,payouts_enabled=excluded.payouts_enabled,updated_at=timezone('utc',now()) where provider_payment_accounts.stripe_account_id=excluded.stripe_account_id returning internal_user_id,stripe_account_id,details_submitted,charges_enabled,payouts_enabled,updated_at`, actor, value.ID, value.DetailsSubmitted, value.ChargesEnabled, value.PayoutsEnabled).Scan(&account.InternalUserID, &account.StripeAccountID, &account.DetailsSubmitted, &account.ChargesEnabled, &account.PayoutsEnabled, &account.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ProviderAccount{}, ErrConflict
	}
	return account, err
}

func (s sqlStore) ApplyProviderEvent(ctx context.Context, event ProviderEvent) error {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx, `insert into public.stripe_webhook_receipts(stripe_event_id,event_type,provider_object_id,outcome) values($1,$2,$3,'processing') on conflict do nothing`, event.ID, event.Kind, event.ProviderObjectID)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return tx.Commit()
	}
	if event.Kind == EventAccountUpdate {
		result, err = tx.ExecContext(ctx, `update public.provider_payment_accounts set details_submitted=$1,charges_enabled=$2,payouts_enabled=$3,updated_at=timezone('utc',now()) where stripe_account_id=$4`, event.DetailsSubmitted, event.ChargesEnabled, event.PayoutsEnabled, event.AccountID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return ErrNotFound
		}
		_, err = tx.ExecContext(ctx, `update public.stripe_webhook_receipts set outcome='account_updated',processed_at=timezone('utc',now()) where stripe_event_id=$1`, event.ID)
		if err != nil {
			return err
		}
		return tx.Commit()
	}
	order, err := loadOrderForEvent(ctx, tx, event)
	if err != nil {
		return err
	}
	from := order.State
	to, eventType, err := eventTransition(from, event.Kind)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `update public.payment_orders set state=$1,stripe_checkout_session_id=coalesce(nullif($2,''),stripe_checkout_session_id),stripe_payment_intent_id=coalesce(nullif($3,''),stripe_payment_intent_id),stripe_invoice_id=coalesce(nullif($4,''),stripe_invoice_id),paid_at=case when $1='paid' then coalesce(paid_at,timezone('utc',now())) else paid_at end,refunded_at=case when $1='refunded' then coalesce(refunded_at,timezone('utc',now())) else refunded_at end,updated_at=timezone('utc',now()) where id=$5`, to, checkoutID(event), event.PaymentIntentID, event.InvoiceID, order.ID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `insert into public.payment_events(payment_order_id,event_type,from_state,to_state,provider_object_id) values($1,$2,$3,$4,$5)`, order.ID, eventType, from, to, event.ProviderObjectID); err != nil {
		return err
	}
	if event.Kind == EventDisputeOpened || event.Kind == EventDisputeWon || event.Kind == EventDisputeLost {
		closed := any(nil)
		if event.Kind == EventDisputeWon || event.Kind == EventDisputeLost {
			closed = time.Now().UTC()
		}
		_, err = tx.ExecContext(ctx, `insert into public.payment_disputes(stripe_dispute_id,payment_order_id,stripe_charge_id,amount_minor,currency,state,reason,opened_at,closed_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9) on conflict(stripe_dispute_id) do update set state=excluded.state,reason=excluded.reason,closed_at=excluded.closed_at,updated_at=timezone('utc',now())`, event.DisputeID, order.ID, event.ChargeID, event.AmountMinor, event.Currency, event.DisputeState, event.DisputeReason, event.OccurredAt, closed)
		if err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `update public.stripe_webhook_receipts set outcome=$1,processed_at=timezone('utc',now()) where stripe_event_id=$2`, eventType, event.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s sqlStore) PrepareRefund(ctx context.Context, actor, orderID uuid.UUID) (Order, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var moderator bool
	if err := tx.QueryRowContext(ctx, `select exists(select 1 from public.platform_roles where internal_user_id=$1 and role in('moderator','administrator'))`, actor).Scan(&moderator); err != nil {
		return Order{}, err
	}
	if !moderator {
		return Order{}, ErrForbidden
	}
	order, found, err := loadOrderByID(ctx, tx, orderID, true)
	if err != nil || !found {
		if !found && err == nil {
			err = ErrNotFound
		}
		return Order{}, err
	}
	if order.State == StateRefundPending {
		if err := tx.Commit(); err != nil {
			return Order{}, err
		}
		return order, nil
	}
	if order.State != StatePaid && order.State != StateDisputeWon || order.PaymentIntentID == "" {
		return Order{}, ErrConflict
	}
	return order, tx.Commit()
}

func (s sqlStore) AttachRefund(ctx context.Context, actor, orderID uuid.UUID, refund RefundResult) (Order, error) {
	tx, err := s.database.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var from State
	if err := tx.QueryRowContext(ctx, `select state from public.payment_orders where id=$1 for update`, orderID).Scan(&from); errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrNotFound
	} else if err != nil {
		return Order{}, err
	}
	result, err := tx.ExecContext(ctx, `update public.payment_orders set state='refund_pending',stripe_refund_id=$1,updated_at=timezone('utc',now()) where id=$2 and state=$3`, refund.ID, orderID, from)
	if err != nil {
		return Order{}, err
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return Order{}, ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `insert into public.payment_events(payment_order_id,event_type,from_state,to_state,provider_object_id) values($1,'refund_requested',$2,'refund_pending',$3)`, orderID, from, refund.ID); err != nil {
		return Order{}, err
	}
	order, _, err := loadOrderByID(ctx, tx, orderID, false)
	if err != nil {
		return Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return Order{}, err
	}
	return order, nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func loadOrderByBooking(ctx context.Context, q queryer, bookingID uuid.UUID) (Order, bool, error) {
	var order Order
	err := q.QueryRowContext(ctx, orderSelect+` where booking_id=$1`, bookingID).Scan(orderScan(&order)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, false, nil
	}
	return order, err == nil, err
}
func loadOrderByID(ctx context.Context, q queryer, orderID uuid.UUID, lock bool) (Order, bool, error) {
	suffix := ` where id=$1`
	if lock {
		suffix += ` for update`
	}
	var order Order
	err := q.QueryRowContext(ctx, orderSelect+suffix, orderID).Scan(orderScan(&order)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, false, nil
	}
	return order, err == nil, err
}

const orderSelect = `select id,booking_id,customer_internal_user_id,provider_internal_user_id,state,gross_minor,platform_fee_minor,provider_net_minor,currency,coalesce(stripe_checkout_session_id,''),coalesce(stripe_payment_intent_id,''),coalesce(stripe_invoice_id,''),coalesce(stripe_refund_id,''),created_at,updated_at from public.payment_orders`

func orderScan(order *Order) []any {
	return []any{&order.ID, &order.BookingID, &order.CustomerID, &order.ProviderID, &order.State, &order.GrossMinor, &order.PlatformFeeMinor, &order.ProviderNetMinor, &order.Currency, &order.CheckoutSessionID, &order.PaymentIntentID, &order.InvoiceID, &order.RefundID, &order.CreatedAt, &order.UpdatedAt}
}
func loadOrderForEvent(ctx context.Context, tx *sql.Tx, event ProviderEvent) (Order, error) {
	var order Order
	var row *sql.Row
	if event.OrderID != "" {
		row = tx.QueryRowContext(ctx, orderSelect+` where id=$1 for update`, event.OrderID)
	} else if event.PaymentIntentID != "" {
		row = tx.QueryRowContext(ctx, orderSelect+` where stripe_payment_intent_id=$1 for update`, event.PaymentIntentID)
	} else {
		return Order{}, ErrInvalid
	}
	if err := row.Scan(orderScan(&order)...); errors.Is(err, sql.ErrNoRows) {
		return Order{}, ErrNotFound
	} else if err != nil {
		return Order{}, err
	}
	return order, nil
}
func eventTransition(from State, kind EventKind) (State, string, error) {
	switch kind {
	case EventProcessing:
		if from == StateCheckoutCreated {
			return StateProcessing, "processing", nil
		}
	case EventPaid:
		if from == StateCheckoutCreated || from == StateProcessing {
			return StatePaid, "paid", nil
		}
		if from == StatePaid {
			return from, "paid", nil
		}
	case EventFailed:
		if from == StateCheckoutCreated || from == StateProcessing {
			return StateFailed, "failed", nil
		}
	case EventRefunded:
		if from == StateRefundPending || from == StatePaid || from == StateDisputed || from == StateDisputeLost {
			return StateRefunded, "refunded", nil
		}
	case EventDisputeOpened:
		if from == StatePaid {
			return StateDisputed, "dispute_opened", nil
		}
	case EventDisputeWon:
		if from == StateDisputed {
			return StateDisputeWon, "dispute_won", nil
		}
	case EventDisputeLost:
		if from == StateDisputed {
			return StateDisputeLost, "dispute_lost", nil
		}
	}
	return "", "", fmt.Errorf("%w: transition %s from %s", ErrConflict, kind, from)
}
func checkoutID(event ProviderEvent) string {
	if len(event.ProviderObjectID) > 3 && event.ProviderObjectID[:3] == "cs_" {
		return event.ProviderObjectID
	}
	return ""
}
