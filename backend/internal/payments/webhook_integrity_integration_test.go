package payments

import (
	"context"
	"errors"
	"testing"
)

func TestWebhookIntegrityRejectsMismatchedFinancialIdentity(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*ProviderEvent)
	}{
		{"checkout_session", func(e *ProviderEvent) { e.ProviderObjectID = "cs_wrong" }},
		{"payment_intent", func(e *ProviderEvent) { e.PaymentIntentID = "pi_wrong" }},
		{"amount", func(e *ProviderEvent) { e.AmountMinor = 1 }},
		{"missing_amount", func(e *ProviderEvent) { e.AmountMinor = 0 }},
		{"currency", func(e *ProviderEvent) { e.Currency = "USD" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPaymentFixture(t, StateCheckoutCreated)
			event := f.event(EventPaid)
			tc.mutate(&event)
			if err := f.store.ApplyProviderEvent(context.Background(), event); err == nil {
				t.Fatal("mismatched event accepted")
			}
			var state State
			var receipts int
			if err := f.db.QueryRow(`select state from public.payment_orders where id=$1`, f.order).Scan(&state); err != nil {
				t.Fatal(err)
			}
			if err := f.db.QueryRow(`select count(*) from public.stripe_webhook_receipts where stripe_event_id=$1`, event.ID).Scan(&receipts); err != nil {
				t.Fatal(err)
			}
			if state != StateCheckoutCreated || receipts != 0 {
				t.Fatalf("rejected event persisted: state=%s receipts=%d", state, receipts)
			}
		})
	}
}

func TestWebhookIntegrityPartialRefundDoesNotMarkFullyRefunded(t *testing.T) {
	f := newPaymentFixture(t, StatePaid)
	e := f.event(EventRefunded)
	e.ProviderObjectID, e.OrderID, e.AmountMinor = "ch_"+f.suffix, "", 100
	if err := f.store.ApplyProviderEvent(context.Background(), e); !errors.Is(err, ErrConflict) {
		t.Fatalf("partial refund marked as full: %v", err)
	}
}

func TestWebhookIntegrityDifferentEventsDoNotDuplicatePaidAudit(t *testing.T) {
	f := newPaymentFixture(t, StateCheckoutCreated)
	e := f.event(EventPaid)
	for _, suffix := range []string{"", "duplicate"} {
		e.ID = "evt_" + f.suffix + suffix
		if err := f.store.ApplyProviderEvent(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := f.db.QueryRow(`select count(*) from public.payment_events where payment_order_id=$1 and event_type='paid'`, f.order).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("same payment produced %d paid events", n)
	}
}

func TestWebhookIntegrityContradictoryReceiptConflicts(t *testing.T) {
	f := newPaymentFixture(t, StateCheckoutCreated)
	e := f.event(EventPaid)
	if err := f.store.ApplyProviderEvent(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	e.Kind = EventFailed
	if err := f.store.ApplyProviderEvent(context.Background(), e); !errors.Is(err, ErrConflict) {
		t.Fatalf("contradictory replay returned %v", err)
	}
}
