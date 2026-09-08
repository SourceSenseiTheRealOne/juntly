package payments

import (
	"context"
	"testing"
)

func TestWebhookIntegritySettledTransitionReplays(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state State
		kind  EventKind
	}{
		{"processing", StateCheckoutCreated, EventProcessing},
		{"failed", StateCheckoutCreated, EventFailed},
		{"refunded", StatePaid, EventRefunded},
		{"refund_after_dispute_won", StateDisputeWon, EventRefunded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPaymentFixture(t, tc.state)
			e := f.event(tc.kind)
			if tc.kind == EventRefunded {
				e.ProviderObjectID = "ch_" + f.suffix
				e.OrderID = ""
			}
			for _, suffix := range []string{"", "", "second"} {
				e.ID = "evt_" + f.suffix + suffix
				if err := f.store.ApplyProviderEvent(context.Background(), e); err != nil {
					t.Fatalf("replay %q: %v", suffix, err)
				}
			}
			var n int
			if err := f.db.QueryRow(`select count(*) from public.payment_events where payment_order_id=$1`, f.order).Scan(&n); err != nil {
				t.Fatal(err)
			}
			if n != 1 {
				t.Fatalf("transition recorded %d times", n)
			}
		})
	}
}
