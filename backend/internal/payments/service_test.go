package payments

import (
	"context"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceBeginsCheckoutFromDurableServerAmounts(t *testing.T) {
	actor := users.InternalUser{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}
	store := &recordingStore{
		order:   Order{ID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", BookingID: "cccccccc-cccc-4ccc-8ccc-cccccccccccc", CustomerID: actor.ID.String(), ProviderID: "dddddddd-dddd-4ddd-8ddd-dddddddddddd", State: StatePendingCheckout, GrossMinor: 12500, PlatformFeeMinor: 1250, ProviderNetMinor: 11250, Currency: "EUR"},
		account: ProviderAccount{StripeAccountID: "acct_provider", DetailsSubmitted: true, ChargesEnabled: true, PayoutsEnabled: true},
	}
	gateway := &recordingGateway{checkout: CheckoutSession{ID: "cs_test_checkout", URL: "https://checkout.stripe.test/session"}}
	service := NewService(staticIdentity{user: actor}, nil, store, gateway, 1000)

	result, err := service.BeginCheckout(context.Background(), users.VerifiedIdentity{Subject: "customer"}, uuid.MustParse(store.order.BookingID), "checkout-key-123", "pt-PT")
	if err != nil {
		t.Fatalf("begin checkout: %v", err)
	}
	if result.URL != gateway.checkout.URL || gateway.checkoutInput.GrossMinor != 12500 || gateway.checkoutInput.FeeMinor != 1250 || gateway.checkoutInput.ConnectedAccountID != "acct_provider" {
		t.Fatalf("checkout result=%#v input=%#v", result, gateway.checkoutInput)
	}
	if store.attachedSession.ID != "cs_test_checkout" {
		t.Fatalf("session not attached: %#v", store.attachedSession)
	}
}

func TestServiceVerifiesWebhookBeforeStoreMutation(t *testing.T) {
	store := &recordingStore{}
	gateway := &recordingGateway{event: ProviderEvent{ID: "evt_paid", Kind: EventPaid, OrderID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", ProviderObjectID: "cs_paid"}}
	service := NewService(staticIdentity{}, nil, store, gateway, 1000)
	if err := service.HandleWebhook(context.Background(), []byte("signed"), "t=1,v1=synthetic"); err != nil {
		t.Fatalf("webhook: %v", err)
	}
	if store.event.ID != "evt_paid" || string(gateway.payload) != "signed" {
		t.Fatalf("event=%#v payload=%q", store.event, gateway.payload)
	}
}

func TestServiceListsAllOrdersOnlyThroughModeratorBoundary(t *testing.T) {
	moderator := users.InternalUser{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}
	store := &recordingStore{order: Order{ID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", State: StateDisputed}}
	service := NewService(staticIdentity{}, staticModerator{user: moderator}, store, nil, 1000)
	orders, err := service.ListAdminOrders(context.Background(), users.VerifiedIdentity{Subject: "moderator"})
	if err != nil || len(orders) != 1 || store.adminActor != moderator.ID {
		t.Fatalf("orders=%#v actor=%s err=%v", orders, store.adminActor, err)
	}
}

func TestServiceUsesOneOrderOwnedRefundKeyAndSkipsPendingReplay(t *testing.T) {
	moderator := users.InternalUser{ID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")}
	orderID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	store := &recordingStore{order: Order{ID: orderID, State: StatePaid, PaymentIntentID: "pi_refund"}}
	gateway := &recordingGateway{}
	service := NewService(staticIdentity{}, staticModerator{user: moderator}, store, gateway, 1000)
	if _, err := service.Refund(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, uuid.MustParse(orderID), "browser-key-one"); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if gateway.refundKey != "refund-"+orderID || gateway.refundCalls != 1 {
		t.Fatalf("refund key=%q calls=%d", gateway.refundKey, gateway.refundCalls)
	}
	if _, err := service.Refund(context.Background(), users.VerifiedIdentity{Subject: "moderator"}, uuid.MustParse(orderID), "browser-key-two"); err != nil {
		t.Fatalf("refund replay: %v", err)
	}
	if gateway.refundCalls != 1 {
		t.Fatalf("pending refund called provider again: %d", gateway.refundCalls)
	}
}

type staticIdentity struct{ user users.InternalUser }

func (s staticIdentity) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	return s.user, false, nil
}

type staticModerator struct{ user users.InternalUser }

func (s staticModerator) RequireModerator(context.Context, users.VerifiedIdentity) (users.InternalUser, error) {
	return s.user, nil
}

type recordingStore struct {
	order           Order
	account         ProviderAccount
	attachedSession CheckoutSession
	event           ProviderEvent
	adminActor      uuid.UUID
}

func (s *recordingStore) PrepareCheckout(context.Context, uuid.UUID, uuid.UUID, string, int) (Order, ProviderAccount, error) {
	return s.order, s.account, nil
}
func (s *recordingStore) AttachCheckout(_ context.Context, _ uuid.UUID, _ uuid.UUID, session CheckoutSession) (Order, error) {
	s.attachedSession = session
	s.order.CheckoutSessionID = session.ID
	s.order.State = StateCheckoutCreated
	return s.order, nil
}
func (s *recordingStore) ListOrders(context.Context, uuid.UUID) ([]Order, error) {
	return []Order{s.order}, nil
}
func (s *recordingStore) ListAdminOrders(_ context.Context, actor uuid.UUID) ([]Order, error) {
	s.adminActor = actor
	return []Order{s.order}, nil
}
func (s *recordingStore) GetProviderAccount(context.Context, uuid.UUID) (ProviderAccount, error) {
	return s.account, nil
}
func (s *recordingStore) SaveProviderAccount(_ context.Context, _ uuid.UUID, account ConnectedAccount) (ProviderAccount, error) {
	s.account = ProviderAccount{StripeAccountID: account.ID, DetailsSubmitted: account.DetailsSubmitted, ChargesEnabled: account.ChargesEnabled, PayoutsEnabled: account.PayoutsEnabled}
	return s.account, nil
}
func (s *recordingStore) ApplyProviderEvent(_ context.Context, event ProviderEvent) error {
	s.event = event
	return nil
}
func (s *recordingStore) PrepareRefund(context.Context, uuid.UUID, uuid.UUID) (Order, error) {
	return s.order, nil
}
func (s *recordingStore) AttachRefund(_ context.Context, _ uuid.UUID, _ uuid.UUID, refund RefundResult) (Order, error) {
	s.order.RefundID = refund.ID
	s.order.State = StateRefundPending
	return s.order, nil
}

type recordingGateway struct {
	checkout      CheckoutSession
	checkoutInput CheckoutRequest
	event         ProviderEvent
	payload       []byte
	refundKey     string
	refundCalls   int
}

func (g *recordingGateway) CreateCheckout(_ context.Context, input CheckoutRequest) (CheckoutSession, error) {
	g.checkoutInput = input
	return g.checkout, nil
}
func (g *recordingGateway) CreateConnectedAccount(context.Context, string) (ConnectedAccount, error) {
	return ConnectedAccount{ID: "acct_provider"}, nil
}
func (g *recordingGateway) CreateAccountLink(context.Context, string, string) (AccountLink, error) {
	return AccountLink{URL: "https://connect.stripe.test/onboarding"}, nil
}
func (g *recordingGateway) GetConnectedAccount(context.Context, string) (ConnectedAccount, error) {
	return ConnectedAccount{ID: "acct_provider"}, nil
}
func (g *recordingGateway) CreateRefund(_ context.Context, _ string, key string) (RefundResult, error) {
	g.refundKey = key
	g.refundCalls++
	return RefundResult{ID: "re_test"}, nil
}
func (g *recordingGateway) VerifyWebhook(payload []byte, _ string) (ProviderEvent, error) {
	g.payload = append([]byte(nil), payload...)
	return g.event, nil
}
