package payments

import "testing"

func TestCheckoutProjectionIncludesReconciliationMoney(t *testing.T) {
	e, err := projectStripeEvent([]byte(`{"id":"evt_reconciliation","type":"checkout.session.completed","created":1800000000,"data":{"object":{"id":"cs_reconciliation","payment_status":"paid","payment_intent":"pi_reconciliation","amount_total":12500,"currency":"eur","metadata":{"order_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}}}}`))
	if err != nil || e.AmountMinor != 12500 || e.Currency != "EUR" {
		t.Fatalf("missing reconciliation money: amount=%d currency=%s err=%v", e.AmountMinor, e.Currency, err)
	}
}
