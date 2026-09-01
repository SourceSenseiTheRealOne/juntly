package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/entitlements"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type EntitlementService interface {
	Catalog(context.Context) (entitlements.Catalog, error)
	RequestSubscription(context.Context, users.VerifiedIdentity, uuid.UUID) (entitlements.Subscription, error)
	CurrentSubscription(context.Context, users.VerifiedIdentity) (*entitlements.Subscription, error)
	RequestPromotion(context.Context, users.VerifiedIdentity, uuid.UUID, uuid.UUID) (entitlements.Promotion, error)
	ListPromotions(context.Context, users.VerifiedIdentity) ([]entitlements.Promotion, error)
	Access(context.Context, users.VerifiedIdentity) (entitlements.Access, error)
}
type entitlementHandler struct{ service EntitlementService }

func NewEntitlementHandler(s EntitlementService) http.Handler { return entitlementHandler{service: s} }

type planRequest struct {
	PlanID *uuid.UUID `json:"planId"`
}
type promotionRequest struct {
	ListingID *uuid.UUID `json:"listingId"`
	PeriodID  *uuid.UUID `json:"periodId"`
}

func (h entitlementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	if r.URL.Path == "/api/v1/entitlements/catalog" {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		v, err := h.service.Catalog(r.Context())
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		writeJSON(w, 200, v, id)
		return
	}
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	switch r.URL.Path {
	case "/api/v1/me/entitlements":
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		access, err := h.service.Access(r.Context(), identity)
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		subscription, err := h.service.CurrentSubscription(r.Context(), identity)
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		promotions, err := h.service.ListPromotions(r.Context(), identity)
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		writeJSON(w, 200, map[string]any{"access": access, "subscription": subscription, "promotions": promotions}, id)
	case "/api/v1/me/subscriptions":
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		var body planRequest
		if !decodeEntitlement(r.Body, &body) || body.PlanID == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		v, err := h.service.RequestSubscription(r.Context(), identity, *body.PlanID)
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		writeJSON(w, 201, v, id)
	case "/api/v1/me/promotions":
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		var body promotionRequest
		if !decodeEntitlement(r.Body, &body) || body.ListingID == nil || body.PeriodID == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		v, err := h.service.RequestPromotion(r.Context(), identity, *body.ListingID, *body.PeriodID)
		if err != nil {
			writeEntitlementError(w, err, id)
			return
		}
		writeJSON(w, 201, v, id)
	default:
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	}
}
func decodeEntitlement(r io.Reader, target any) bool {
	d := json.NewDecoder(io.LimitReader(r, 4097))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(d.Decode(&extra), io.EOF)
}
func writeEntitlementError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, entitlements.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, entitlements.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, entitlements.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, entitlements.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
