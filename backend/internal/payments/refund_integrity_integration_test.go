package payments

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type paymentFixture struct {
	db                        *sql.DB
	store                     Store
	customer, provider, order uuid.UUID
	suffix                    string
}

func newPaymentFixture(t *testing.T, state State) paymentFixture {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	f := paymentFixture{db: db, store: NewSQLStore(db), customer: uuid.New(), provider: uuid.New(), order: uuid.New(), suffix: strings.ReplaceAll(uuid.NewString(), "-", "")}
	booking := uuid.New()
	t.Cleanup(func() {
		defer db.Close()
		for _, q := range []string{
			`delete from public.stripe_webhook_receipts where stripe_event_id like 'evt_' || $1 || '%'`,
		} {
			if _, err := db.Exec(q, f.suffix); err != nil {
				t.Error(err)
			}
		}
		for _, q := range []string{
			`delete from public.payment_orders where id=$1`,
		} {
			if _, err := db.Exec(q, f.order); err != nil {
				t.Error(err)
			}
		}
		if _, err := db.Exec(`delete from public.bookings where id=$1`, booking); err != nil {
			t.Error(err)
		}
		if _, err := db.Exec(`delete from public.provider_profiles where internal_user_id=$1`, f.provider); err != nil {
			t.Error(err)
		}
		if _, err := db.Exec(`delete from public.platform_roles where internal_user_id=$1`, f.provider); err != nil {
			t.Error(err)
		}
		if _, err := db.Exec(`delete from public.user_accounts where internal_user_id in($1,$2)`, f.customer, f.provider); err != nil {
			t.Error(err)
		}
		if _, err := db.Exec(`delete from public.internal_users where id in($1,$2)`, f.customer, f.provider); err != nil {
			t.Error(err)
		}
	})
	for _, id := range []uuid.UUID{f.customer, f.provider} {
		if _, err := db.Exec(`insert into public.internal_users(id,clerk_subject) values($1,$2)`, id, "synthetic_"+id.String()); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`insert into public.user_accounts(internal_user_id,provider_enabled) values($1,$2)`, id, id == f.provider); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`insert into public.provider_profiles(internal_user_id,display_name,provider_type,bio,primary_locality_id,max_travel_distance_km,remote_services) select $1,'Synthetic provider','professional','Isolated payment regression fixture.',id,0,true from public.localities order by id limit 1`, f.provider); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`insert into public.platform_roles(id,internal_user_id,role) values($1,$2,'moderator')`, uuid.New(), f.provider); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`insert into public.bookings(id,customer_internal_user_id,provider_internal_user_id,source_type,idempotency_key,state,scheduled_at,private_location,agreed_price_minor) values($1,$2,$3,'direct',$4,'confirmed',now()+interval '1 day','Synthetic test location',12500)`, booking, f.customer, f.provider, f.suffix); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`insert into public.payment_orders(id,booking_id,customer_internal_user_id,provider_internal_user_id,idempotency_key,state,gross_minor,platform_fee_minor,provider_net_minor,currency,stripe_checkout_session_id,stripe_payment_intent_id) values($1,$2,$3,$4,$5,$6,12500,1250,11250,'EUR',$7,$8)`, f.order, booking, f.customer, f.provider, f.suffix, state, "cs_"+f.suffix, "pi_"+f.suffix); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f paymentFixture) event(kind EventKind) ProviderEvent {
	return ProviderEvent{ID: "evt_" + f.suffix, Kind: kind, ProviderObjectID: "cs_" + f.suffix, OrderID: f.order.String(), PaymentIntentID: "pi_" + f.suffix, AmountMinor: 12500, Currency: "EUR"}
}

func TestRefundIntegrityWebhookWinsRace(t *testing.T) {
	f := newPaymentFixture(t, StatePaid)
	event := f.event(EventRefunded)
	event.ProviderObjectID, event.OrderID = "ch_"+f.suffix, ""
	if err := f.store.ApplyProviderEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		order, err := f.store.AttachRefund(context.Background(), f.provider, f.order, RefundResult{ID: "re_" + f.suffix})
		if err != nil {
			t.Fatal(err)
		}
		if order.State != StateRefunded {
			t.Fatalf("late refund response regressed terminal state to %s", order.State)
		}
	}
	var count int
	if err := f.db.QueryRow(`select count(*) from public.payment_events where payment_order_id=$1 and event_type='refund_requested'`, f.order).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("late response wrote %d redundant events", count)
	}
}

func TestRefundIntegrityRejectsUnauthorizedAttachment(t *testing.T) {
	f := newPaymentFixture(t, StatePaid)
	_, err := f.store.AttachRefund(context.Background(), f.customer, f.order, RefundResult{ID: "re_" + f.suffix})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("unauthorized attachment: %v", err)
	}
}

func TestRefundIntegrityReplayDoesNotDuplicateAudit(t *testing.T) {
	f := newPaymentFixture(t, StatePaid)
	for range 2 {
		if _, err := f.store.AttachRefund(context.Background(), f.provider, f.order, RefundResult{ID: "re_" + f.suffix}); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := f.db.QueryRow(`select count(*) from public.payment_events where payment_order_id=$1 and event_type='refund_requested'`, f.order).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("refund replay wrote %d audit events, want 1", count)
	}
	if _, err := f.store.AttachRefund(context.Background(), f.provider, f.order, RefundResult{ID: "re_other"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("different refund replaced pending refund: %v", err)
	}
}

func TestRefundIntegrityRejectsIneligibleState(t *testing.T) {
	f := newPaymentFixture(t, StateDisputeLost)
	if _, err := f.store.AttachRefund(context.Background(), f.provider, f.order, RefundResult{ID: "re_" + f.suffix}); !errors.Is(err, ErrConflict) {
		t.Fatalf("ineligible attachment: %v", err)
	}
}
