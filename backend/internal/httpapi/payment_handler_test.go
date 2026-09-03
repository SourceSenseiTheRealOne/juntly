package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/payments"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestPaymentHandlerRoutesCheckoutWithoutAcceptingMoneyFromBrowser(t *testing.T) {
	service := &recordingPaymentService{checkout: payments.CheckoutResult{Order: payments.Order{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", GrossMinor: 12500, PlatformFeeMinor: 1250, ProviderNetMinor: 11250, Currency: "EUR"}, URL: "https://checkout.stripe.test/session"}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "customer"}}, httpapi.NewPaymentHandler(service))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/bookings/bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb/checkout", strings.NewReader(`{"idempotencyKey":"checkout-key-123","locale":"pt-PT"}`))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || service.bookingID.String() != "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" || service.idempotencyKey != "checkout-key-123" || strings.Contains(response.Body.String(), "stripeAccount") {
		t.Fatalf("response=%d/%s service=%#v", response.Code, response.Body.String(), service)
	}
}

func TestStripeWebhookHandlerPassesExactRawBodyAndSignature(t *testing.T) {
	service := &recordingPaymentService{}
	handler := httpapi.NewStripeWebhookHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/webhooks/stripe", strings.NewReader(`{"id":"evt_test"}`))
	request.Header.Set("Stripe-Signature", "t=1,v1=synthetic")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || string(service.payload) != `{"id":"evt_test"}` || service.signature != "t=1,v1=synthetic" {
		t.Fatalf("response=%d body=%q signature=%q", response.Code, service.payload, service.signature)
	}
}

func TestPaymentHandlerListsAdministratorPaymentOrders(t *testing.T) {
	service := &recordingPaymentService{orders: []payments.Order{{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", State: payments.StateDisputed}}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "moderator"}}, httpapi.NewPaymentHandler(service))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/payments", nil)
	request.Header.Set("Authorization", "Bearer synthetic-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "disputed") || service.adminListCalls != 1 {
		t.Fatalf("response=%d/%s calls=%d", response.Code, response.Body.String(), service.adminListCalls)
	}
}

type recordingPaymentService struct {
	checkout       payments.CheckoutResult
	bookingID      uuid.UUID
	idempotencyKey string
	payload        []byte
	signature      string
	orders         []payments.Order
	adminListCalls int
}

func (s *recordingPaymentService) BeginCheckout(_ context.Context, _ users.VerifiedIdentity, bookingID uuid.UUID, key, _ string) (payments.CheckoutResult, error) {
	s.bookingID, s.idempotencyKey = bookingID, key
	return s.checkout, nil
}
func (s *recordingPaymentService) ListOrders(context.Context, users.VerifiedIdentity) ([]payments.Order, error) {
	return nil, nil
}
func (s *recordingPaymentService) ListAdminOrders(context.Context, users.VerifiedIdentity) ([]payments.Order, error) {
	s.adminListCalls++
	return s.orders, nil
}
func (s *recordingPaymentService) BeginPayoutOnboarding(context.Context, users.VerifiedIdentity, string) (payments.PayoutOnboardingResult, error) {
	return payments.PayoutOnboardingResult{}, nil
}
func (s *recordingPaymentService) PayoutStatus(context.Context, users.VerifiedIdentity) (payments.ProviderAccount, error) {
	return payments.ProviderAccount{}, nil
}
func (s *recordingPaymentService) HandleWebhook(_ context.Context, payload []byte, signature string) error {
	s.payload, s.signature = append([]byte(nil), payload...), signature
	return nil
}
func (s *recordingPaymentService) Refund(context.Context, users.VerifiedIdentity, uuid.UUID, string) (payments.Order, error) {
	return payments.Order{}, nil
}
