package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	stripeBodyLimit = 256 * 1024
	stripeTolerance = 5 * time.Minute
)

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	APIBase       string
	PublicOrigin  string
	HTTPClient    *http.Client
	Now           func() time.Time
}

type StripeGateway struct {
	secretKey     string
	webhookSecret string
	apiBase       string
	publicOrigin  string
	client        *http.Client
	now           func() time.Time
}

func NewStripeGateway(config StripeConfig) (*StripeGateway, error) {
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	config.WebhookSecret = strings.TrimSpace(config.WebhookSecret)
	config.APIBase = strings.TrimRight(strings.TrimSpace(config.APIBase), "/")
	config.PublicOrigin = strings.TrimRight(strings.TrimSpace(config.PublicOrigin), "/")
	api, apiErr := url.Parse(config.APIBase)
	public, publicErr := url.Parse(config.PublicOrigin)
	if !strings.HasPrefix(config.SecretKey, "sk_") || !strings.HasPrefix(config.WebhookSecret, "whsec_") || apiErr != nil || api.Host == "" || (api.Scheme != "https" && api.Hostname() != "127.0.0.1" && api.Hostname() != "localhost") || publicErr != nil || public.Scheme != "https" || public.Host == "" || public.Path != "" {
		return nil, ErrUnavailable
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	return &StripeGateway{secretKey: config.SecretKey, webhookSecret: config.WebhookSecret, apiBase: config.APIBase, publicOrigin: config.PublicOrigin, client: config.HTTPClient, now: config.Now}, nil
}

func (g *StripeGateway) CreateCheckout(ctx context.Context, input CheckoutRequest) (CheckoutSession, error) {
	if input.OrderID == "" || input.BookingID == "" || !strings.HasPrefix(input.ConnectedAccountID, "acct_") || input.GrossMinor < 50 || input.FeeMinor < 0 || input.FeeMinor >= input.GrossMinor || len(input.IdempotencyKey) < 8 {
		return CheckoutSession{}, ErrInvalid
	}
	locale := input.Locale
	if locale != "pt-PT" && locale != "en" && locale != "es" {
		locale = "pt-PT"
	}
	values := url.Values{
		"mode":                                   {"payment"},
		"success_url":                            {g.publicOrigin + "/" + locale + "/account/bookings?payment=returned"},
		"cancel_url":                             {g.publicOrigin + "/" + locale + "/account/bookings?payment=cancelled"},
		"line_items[0][price_data][currency]":    {"eur"},
		"line_items[0][price_data][unit_amount]": {strconv.FormatInt(input.GrossMinor, 10)},
		"line_items[0][price_data][product_data][name]":     {"Vila service booking"},
		"line_items[0][quantity]":                           {"1"},
		"payment_intent_data[application_fee_amount]":       {strconv.FormatInt(input.FeeMinor, 10)},
		"payment_intent_data[transfer_data][destination]":   {input.ConnectedAccountID},
		"metadata[order_id]":                                {input.OrderID},
		"metadata[booking_id]":                              {input.BookingID},
		"automatic_tax[enabled]":                            {"true"},
		"tax_id_collection[enabled]":                        {"true"},
		"invoice_creation[enabled]":                         {"true"},
		"billing_address_collection":                        {"required"},
		"customer_creation":                                 {"always"},
		"payment_method_types[0]":                           {"card"},
		"payment_method_types[1]":                           {"mb_way"},
		"consent_collection[terms_of_service]":              {"required"},
		"custom_text[terms_of_service_acceptance][message]": {"I agree to Vila's terms, payment policy, and refund policy."},
	}
	var response struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := g.form(ctx, http.MethodPost, "/v1/checkout/sessions", values, input.IdempotencyKey, &response); err != nil || !strings.HasPrefix(response.ID, "cs_") || !validHTTPS(response.URL) {
		return CheckoutSession{}, ErrUnavailable
	}
	return CheckoutSession{ID: response.ID, URL: response.URL}, nil
}

func (g *StripeGateway) CreateConnectedAccount(ctx context.Context, idempotencyKey string) (ConnectedAccount, error) {
	values := url.Values{"type": {"express"}, "country": {"PT"}, "capabilities[card_payments][requested]": {"true"}, "capabilities[transfers][requested]": {"true"}, "business_type": {"individual"}}
	var response stripeAccount
	if err := g.form(ctx, http.MethodPost, "/v1/accounts", values, idempotencyKey, &response); err != nil || !strings.HasPrefix(response.ID, "acct_") {
		return ConnectedAccount{}, ErrUnavailable
	}
	return response.account(), nil
}

func (g *StripeGateway) CreateAccountLink(ctx context.Context, accountID, locale string) (AccountLink, error) {
	if !strings.HasPrefix(accountID, "acct_") {
		return AccountLink{}, ErrInvalid
	}
	values := url.Values{"account": {accountID}, "type": {"account_onboarding"}, "refresh_url": {g.publicOrigin + "/" + locale + "/account/payouts?onboarding=refresh"}, "return_url": {g.publicOrigin + "/" + locale + "/account/payouts?onboarding=returned"}}
	var response struct {
		URL string `json:"url"`
	}
	if err := g.form(ctx, http.MethodPost, "/v1/account_links", values, "", &response); err != nil || !validHTTPS(response.URL) {
		return AccountLink{}, ErrUnavailable
	}
	return AccountLink{URL: response.URL}, nil
}

func (g *StripeGateway) GetConnectedAccount(ctx context.Context, accountID string) (ConnectedAccount, error) {
	if !strings.HasPrefix(accountID, "acct_") {
		return ConnectedAccount{}, ErrInvalid
	}
	var response stripeAccount
	if err := g.form(ctx, http.MethodGet, "/v1/accounts/"+url.PathEscape(accountID), nil, "", &response); err != nil {
		return ConnectedAccount{}, ErrUnavailable
	}
	return response.account(), nil
}

func (g *StripeGateway) CreateRefund(ctx context.Context, paymentIntentID, idempotencyKey string) (RefundResult, error) {
	if !strings.HasPrefix(paymentIntentID, "pi_") || len(idempotencyKey) < 8 {
		return RefundResult{}, ErrInvalid
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := g.form(ctx, http.MethodPost, "/v1/refunds", url.Values{"payment_intent": {paymentIntentID}, "reverse_transfer": {"true"}, "refund_application_fee": {"true"}}, idempotencyKey, &response); err != nil || !strings.HasPrefix(response.ID, "re_") {
		return RefundResult{}, ErrUnavailable
	}
	return RefundResult{ID: response.ID}, nil
}

func (g *StripeGateway) VerifyWebhook(payload []byte, signature string) (ProviderEvent, error) {
	if len(payload) == 0 || len(payload) > stripeBodyLimit {
		return ProviderEvent{}, ErrInvalid
	}
	timestamp, signatures, ok := parseStripeSignature(signature)
	if !ok || g.now().Sub(time.Unix(timestamp, 0)).Abs() > stripeTolerance {
		return ProviderEvent{}, ErrUnauthorized
	}
	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := mac.Sum(nil)
	verified := false
	for _, candidate := range signatures {
		decoded, err := hex.DecodeString(candidate)
		if err == nil && hmac.Equal(decoded, expected) {
			verified = true
			break
		}
	}
	if !verified {
		return ProviderEvent{}, ErrUnauthorized
	}
	return projectStripeEvent(payload)
}

type stripeAccount struct {
	ID               string `json:"id"`
	DetailsSubmitted bool   `json:"details_submitted"`
	ChargesEnabled   bool   `json:"charges_enabled"`
	PayoutsEnabled   bool   `json:"payouts_enabled"`
}

func (a stripeAccount) account() ConnectedAccount {
	return ConnectedAccount{ID: a.ID, DetailsSubmitted: a.DetailsSubmitted, ChargesEnabled: a.ChargesEnabled, PayoutsEnabled: a.PayoutsEnabled}
}

func (g *StripeGateway) form(ctx context.Context, method, path string, values url.Values, idempotencyKey string, target any) error {
	var body io.Reader
	if values != nil {
		body = strings.NewReader(values.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, g.apiBase+path, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(g.secretKey, "")
	if values != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, stripeBodyLimit))
		return fmt.Errorf("stripe status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, stripeBodyLimit))
	return decoder.Decode(target)
}

func parseStripeSignature(value string) (int64, []string, bool) {
	var timestamp int64
	var signatures []string
	for _, part := range strings.Split(value, ",") {
		key, raw, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestamp, _ = strconv.ParseInt(raw, 10, 64)
		case "v1":
			signatures = append(signatures, raw)
		}
	}
	return timestamp, signatures, timestamp > 0 && len(signatures) > 0
}

func projectStripeEvent(payload []byte) (ProviderEvent, error) {
	var envelope struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Created int64  `json:"created"`
		Data    struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil || !strings.HasPrefix(envelope.ID, "evt_") {
		return ProviderEvent{}, ErrInvalid
	}
	event := ProviderEvent{ID: envelope.ID, OccurredAt: time.Unix(envelope.Created, 0).UTC()}
	switch envelope.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.async_payment_failed":
		var object struct {
			ID            string            `json:"id"`
			PaymentStatus string            `json:"payment_status"`
			PaymentIntent string            `json:"payment_intent"`
			Invoice       string            `json:"invoice"`
			Metadata      map[string]string `json:"metadata"`
		}
		if json.Unmarshal(envelope.Data.Object, &object) != nil || !strings.HasPrefix(object.ID, "cs_") || object.Metadata["order_id"] == "" {
			return ProviderEvent{}, ErrInvalid
		}
		event.ProviderObjectID, event.OrderID, event.PaymentIntentID, event.InvoiceID = object.ID, object.Metadata["order_id"], object.PaymentIntent, object.Invoice
		if envelope.Type == "checkout.session.async_payment_failed" {
			event.Kind = EventFailed
		} else if object.PaymentStatus == "paid" || envelope.Type == "checkout.session.async_payment_succeeded" {
			event.Kind = EventPaid
		} else {
			event.Kind = EventProcessing
		}
	case "charge.refunded":
		var object struct {
			ID             string `json:"id"`
			PaymentIntent  string `json:"payment_intent"`
			Currency       string `json:"currency"`
			AmountRefunded int64  `json:"amount_refunded"`
		}
		if json.Unmarshal(envelope.Data.Object, &object) != nil {
			return ProviderEvent{}, ErrInvalid
		}
		event.Kind, event.ProviderObjectID, event.ChargeID, event.PaymentIntentID, event.AmountMinor, event.Currency = EventRefunded, object.ID, object.ID, object.PaymentIntent, object.AmountRefunded, strings.ToUpper(object.Currency)
	case "charge.dispute.created", "charge.dispute.closed":
		var object struct {
			ID            string `json:"id"`
			Charge        string `json:"charge"`
			PaymentIntent string `json:"payment_intent"`
			Status        string `json:"status"`
			Reason        string `json:"reason"`
			Currency      string `json:"currency"`
			Amount        int64  `json:"amount"`
		}
		if json.Unmarshal(envelope.Data.Object, &object) != nil {
			return ProviderEvent{}, ErrInvalid
		}
		event.ProviderObjectID, event.DisputeID, event.ChargeID, event.PaymentIntentID, event.DisputeState, event.DisputeReason, event.AmountMinor, event.Currency = object.ID, object.ID, object.Charge, object.PaymentIntent, object.Status, object.Reason, object.Amount, strings.ToUpper(object.Currency)
		switch object.Status {
		case "won":
			event.Kind = EventDisputeWon
		case "lost", "warning_closed":
			event.Kind = EventDisputeLost
		default:
			event.Kind = EventDisputeOpened
		}
	case "account.updated":
		var object stripeAccount
		if json.Unmarshal(envelope.Data.Object, &object) != nil || !strings.HasPrefix(object.ID, "acct_") {
			return ProviderEvent{}, ErrInvalid
		}
		event.Kind, event.ProviderObjectID, event.AccountID, event.DetailsSubmitted, event.ChargesEnabled, event.PayoutsEnabled = EventAccountUpdate, object.ID, object.ID, object.DetailsSubmitted, object.ChargesEnabled, object.PayoutsEnabled
	default:
		return ProviderEvent{}, ErrInvalid
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Unix(0, 0).UTC()
	}
	return event, nil
}

func validHTTPS(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}
