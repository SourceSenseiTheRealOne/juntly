package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/payments"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxPaymentRequestBytes = 256 * 1024

type PaymentService interface {
	BeginCheckout(context.Context, users.VerifiedIdentity, uuid.UUID, string, string) (payments.CheckoutResult, error)
	ListOrders(context.Context, users.VerifiedIdentity) ([]payments.Order, error)
	ListAdminOrders(context.Context, users.VerifiedIdentity) ([]payments.Order, error)
	BeginPayoutOnboarding(context.Context, users.VerifiedIdentity, string) (payments.PayoutOnboardingResult, error)
	PayoutStatus(context.Context, users.VerifiedIdentity) (payments.ProviderAccount, error)
	HandleWebhook(context.Context, []byte, string) error
	Refund(context.Context, users.VerifiedIdentity, uuid.UUID, string) (payments.Order, error)
}

type paymentHandler struct{ service PaymentService }

type stripeWebhookHandler struct{ service PaymentService }

func NewPaymentHandler(service PaymentService) http.Handler { return paymentHandler{service: service} }
func NewStripeWebhookHandler(service PaymentService) http.Handler {
	return stripeWebhookHandler{service: service}
}

func (h paymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/api/v1/me/payments" && r.Method == http.MethodGet:
		orders, err := h.service.ListOrders(r.Context(), identity)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 200, map[string]any{"orders": orders}, requestID)
	case path == "/api/v1/admin/payments" && r.Method == http.MethodGet:
		orders, err := h.service.ListAdminOrders(r.Context(), identity)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 200, map[string]any{"orders": orders}, requestID)
	case path == "/api/v1/me/payout-account" && r.Method == http.MethodGet:
		account, err := h.service.PayoutStatus(r.Context(), identity)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 200, account, requestID)
	case path == "/api/v1/me/payout-account" && r.Method == http.MethodPost:
		var body struct {
			Locale *string `json:"locale"`
		}
		if !decodePayment(r.Body, &body) || body.Locale == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		result, err := h.service.BeginPayoutOnboarding(r.Context(), identity, *body.Locale)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 200, result, requestID)
	case strings.HasPrefix(path, "/api/v1/me/bookings/") && strings.HasSuffix(path, "/checkout") && r.Method == http.MethodPost:
		id, ok := paymentPathID(path, "/api/v1/me/bookings/", "/checkout")
		if !ok {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		var body struct {
			IdempotencyKey *string `json:"idempotencyKey"`
			Locale         *string `json:"locale"`
		}
		if !decodePayment(r.Body, &body) || body.IdempotencyKey == nil || body.Locale == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		result, err := h.service.BeginCheckout(r.Context(), identity, id, *body.IdempotencyKey, *body.Locale)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 201, result, requestID)
	case strings.HasPrefix(path, "/api/v1/admin/payments/") && strings.HasSuffix(path, "/refund") && r.Method == http.MethodPost:
		id, ok := paymentPathID(path, "/api/v1/admin/payments/", "/refund")
		if !ok {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		var body struct {
			IdempotencyKey *string `json:"idempotencyKey"`
		}
		if !decodePayment(r.Body, &body) || body.IdempotencyKey == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		order, err := h.service.Refund(r.Context(), identity, id, *body.IdempotencyKey)
		if err != nil {
			writePaymentError(w, err, requestID)
			return
		}
		writeJSON(w, 200, order, requestID)
	default:
		writeAPIError(w, 404, "NOT_FOUND", "Not found", requestID)
	}
}

func (h stripeWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(405), 405)
		return
	}
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}
	signature := strings.TrimSpace(r.Header.Get("Stripe-Signature"))
	if signature == "" {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxPaymentRequestBytes+1))
	if err != nil || len(payload) == 0 || len(payload) > maxPaymentRequestBytes {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	if err := h.service.HandleWebhook(r.Context(), payload, signature); err != nil {
		writePaymentError(w, err, requestID)
		return
	}
	writeJSON(w, 200, map[string]bool{"received": true}, requestID)
}

func paymentPathID(path, prefix, suffix string) (uuid.UUID, bool) {
	raw := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	if strings.Contains(raw, "/") {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	return id, err == nil
}

func decodePayment(body io.Reader, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(body, 8*1024+1))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func writePaymentError(w http.ResponseWriter, err error, requestID string) {
	switch {
	case errors.Is(err, payments.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
	case errors.Is(err, payments.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", requestID)
	case errors.Is(err, payments.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", requestID)
	case errors.Is(err, payments.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", requestID)
	case errors.Is(err, payments.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", requestID)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
	}
}
