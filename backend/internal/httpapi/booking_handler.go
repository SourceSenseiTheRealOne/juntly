package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/bookings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxBookingRequestBytes = 8 * 1024

type BookingService interface {
	Create(context.Context, users.VerifiedIdentity, bookings.CreateBooking) (bookings.Booking, error)
	List(context.Context, users.VerifiedIdentity) ([]bookings.Booking, error)
	Get(context.Context, users.VerifiedIdentity, uuid.UUID) (bookings.Booking, error)
	Transition(context.Context, users.VerifiedIdentity, uuid.UUID, bookings.Transition) (bookings.Booking, error)
}
type bookingHandler struct{ service BookingService }

func NewBookingHandler(service BookingService) http.Handler { return bookingHandler{service: service} }

type createBookingRequest struct {
	SourceType       *bookings.SourceType `json:"sourceType"`
	SourceID         *uuid.UUID           `json:"sourceId"`
	ProviderID       *uuid.UUID           `json:"providerId"`
	IdempotencyKey   *string              `json:"idempotencyKey"`
	ScheduledAt      *time.Time           `json:"scheduledAt"`
	PrivateLocation  *string              `json:"privateLocation"`
	AgreedPriceMinor *int                 `json:"agreedPriceMinor"`
}
type transitionBookingRequest struct {
	ExpectedState *bookings.State `json:"expectedState"`
	TargetState   *bookings.State `json:"targetState"`
	Revision      *int            `json:"revision"`
	Reason        *string         `json:"reason"`
}
type bookingResponse struct {
	ID               uuid.UUID           `json:"id"`
	CustomerID       uuid.UUID           `json:"customerId"`
	ProviderID       uuid.UUID           `json:"providerId"`
	SourceType       bookings.SourceType `json:"sourceType"`
	SourceID         *uuid.UUID          `json:"sourceId"`
	State            bookings.State      `json:"state"`
	Revision         int                 `json:"revision"`
	ScheduledAt      string              `json:"scheduledAt"`
	PrivateLocation  string              `json:"privateLocation"`
	AgreedPriceMinor int                 `json:"agreedPriceMinor"`
	Currency         string              `json:"currency"`
	CreatedAt        string              `json:"createdAt"`
	UpdatedAt        string              `json:"updatedAt"`
}

func (h bookingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	base := "/api/v1/me/bookings"
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == base {
		switch r.Method {
		case http.MethodGet:
			values, err := h.service.List(r.Context(), identity)
			if err != nil {
				writeBookingError(w, err, requestID)
				return
			}
			writeJSON(w, 200, map[string]any{"bookings": bookingResponses(values)}, requestID)
		case http.MethodPost:
			var body createBookingRequest
			if !decodeBooking(r.Body, &body) || body.SourceType == nil || body.IdempotencyKey == nil || body.ScheduledAt == nil || body.PrivateLocation == nil {
				writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			value, err := h.service.Create(r.Context(), identity, bookings.CreateBooking{SourceType: *body.SourceType, SourceID: body.SourceID, ProviderID: body.ProviderID, IdempotencyKey: *body.IdempotencyKey, ScheduledAt: *body.ScheduledAt, PrivateLocation: *body.PrivateLocation, AgreedPriceMinor: body.AgreedPriceMinor})
			if err != nil {
				writeBookingError(w, err, requestID)
				return
			}
			writeJSON(w, 201, bookingJSON(value), requestID)
		default:
			http.Error(w, http.StatusText(405), 405)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, base+"/"), "/")
	if len(parts) < 1 {
		writeAPIError(w, 404, "NOT_FOUND", "Not found", requestID)
		return
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		value, err := h.service.Get(r.Context(), identity, id)
		if err != nil {
			writeBookingError(w, err, requestID)
			return
		}
		writeJSON(w, 200, bookingJSON(value), requestID)
		return
	}
	if len(parts) == 2 && parts[1] == "transitions" && r.Method == http.MethodPost {
		var body transitionBookingRequest
		if !decodeBooking(r.Body, &body) || body.ExpectedState == nil || body.TargetState == nil || body.Revision == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		value, err := h.service.Transition(r.Context(), identity, id, bookings.Transition{ExpectedState: *body.ExpectedState, TargetState: *body.TargetState, Revision: *body.Revision, Reason: body.Reason})
		if err != nil {
			writeBookingError(w, err, requestID)
			return
		}
		writeJSON(w, 200, bookingJSON(value), requestID)
		return
	}
	writeAPIError(w, 404, "NOT_FOUND", "Not found", requestID)
}
func decodeBooking(body io.Reader, target any) bool {
	d := json.NewDecoder(io.LimitReader(body, maxBookingRequestBytes+1))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(d.Decode(&extra), io.EOF)
}
func bookingJSON(v bookings.Booking) bookingResponse {
	return bookingResponse{ID: v.ID, CustomerID: v.CustomerID, ProviderID: v.ProviderID, SourceType: v.SourceType, SourceID: v.SourceID, State: v.State, Revision: v.Revision, ScheduledAt: v.ScheduledAt.UTC().Format(time.RFC3339Nano), PrivateLocation: v.PrivateLocation, AgreedPriceMinor: v.AgreedPriceMinor, Currency: v.Currency, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func bookingResponses(values []bookings.Booking) []bookingResponse {
	out := make([]bookingResponse, 0, len(values))
	for _, v := range values {
		out = append(out, bookingJSON(v))
	}
	return out
}
func writeBookingError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, bookings.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, bookings.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, bookings.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, bookings.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	case errors.Is(err, bookings.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
