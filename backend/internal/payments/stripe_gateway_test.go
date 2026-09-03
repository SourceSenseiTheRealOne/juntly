package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestStripeGatewayCreatesServerPricedDestinationCheckout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/checkout/sessions" || r.Method != http.MethodPost {
			t.Fatalf("unexpected Stripe request %s %s", r.Method, r.URL.Path)
		}
		if user, _, ok := r.BasicAuth(); !ok || user != "sk_test_synthetic" {
			t.Fatal("missing Stripe server authentication")
		}
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		for key, expected := range map[string]string{
			"mode":                                            "payment",
			"line_items[0][price_data][currency]":             "eur",
			"line_items[0][price_data][unit_amount]":          "12500",
			"line_items[0][quantity]":                         "1",
			"payment_intent_data[application_fee_amount]":     "1250",
			"payment_intent_data[transfer_data][destination]": "acct_provider",
			"metadata[order_id]":                              "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			"automatic_tax[enabled]":                          "true",
			"tax_id_collection[enabled]":                      "true",
			"invoice_creation[enabled]":                       "true",
		} {
			if values.Get(key) != expected {
				t.Errorf("%s = %q, want %q", key, values.Get(key), expected)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"cs_test_checkout","url":"https://checkout.stripe.test/session"}`))
	}))
	defer server.Close()

	gateway, err := NewStripeGateway(StripeConfig{SecretKey: "sk_test_synthetic", WebhookSecret: "whsec_synthetic", APIBase: server.URL, PublicOrigin: "https://vila.example", Now: time.Now})
	if err != nil {
		t.Fatalf("gateway: %v", err)
	}
	session, err := gateway.CreateCheckout(context.Background(), CheckoutRequest{OrderID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", BookingID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", ConnectedAccountID: "acct_provider", GrossMinor: 12500, FeeMinor: 1250, Locale: "pt-PT", IdempotencyKey: "checkout-key-123"})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if session.ID != "cs_test_checkout" || session.URL != "https://checkout.stripe.test/session" {
		t.Fatalf("session = %#v", session)
	}
}

func TestStripeGatewayVerifiesRawWebhookAndRejectsTampering(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	gateway, err := NewStripeGateway(StripeConfig{SecretKey: "sk_test_synthetic", WebhookSecret: "whsec_synthetic", APIBase: "https://api.stripe.test", PublicOrigin: "https://vila.example", Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"id":"evt_checkout","type":"checkout.session.completed","data":{"object":{"id":"cs_live","payment_status":"paid","payment_intent":"pi_live","invoice":"in_live","metadata":{"order_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}}}}`)
	signed := []byte("1800000000." + string(payload))
	mac := hmac.New(sha256.New, []byte("whsec_synthetic"))
	_, _ = mac.Write(signed)
	signature := "t=1800000000,v1=" + hex.EncodeToString(mac.Sum(nil))

	event, err := gateway.VerifyWebhook(payload, signature)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if event.ID != "evt_checkout" || event.Kind != EventPaid || event.OrderID != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa" || event.PaymentIntentID != "pi_live" || event.InvoiceID != "in_live" {
		t.Fatalf("event = %#v", event)
	}
	if _, err := gateway.VerifyWebhook([]byte(strings.ReplaceAll(string(payload), "paid", "unpaid")), signature); err == nil {
		t.Fatal("tampered payload accepted")
	}
}

func TestProjectStripeRefundAndDisputeFields(t *testing.T) {
	refund, err := projectStripeEvent([]byte(`{"id":"evt_refund","type":"charge.refunded","created":1800000000,"data":{"object":{"id":"ch_refund","payment_intent":"pi_refund","amount_refunded":12500,"currency":"eur"}}}`))
	if err != nil || refund.Kind != EventRefunded || refund.PaymentIntentID != "pi_refund" || refund.AmountMinor != 12500 || refund.Currency != "EUR" {
		t.Fatalf("refund = %#v, err = %v", refund, err)
	}
	dispute, err := projectStripeEvent([]byte(`{"id":"evt_dispute","type":"charge.dispute.created","created":1800000000,"data":{"object":{"id":"dp_test","charge":"ch_test","payment_intent":"pi_test","status":"needs_response","reason":"fraudulent","amount":12500,"currency":"eur"}}}`))
	if err != nil || dispute.Kind != EventDisputeOpened || dispute.PaymentIntentID != "pi_test" || dispute.DisputeID != "dp_test" || dispute.AmountMinor != 12500 {
		t.Fatalf("dispute = %#v, err = %v", dispute, err)
	}
}
