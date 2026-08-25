package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/moderation"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

type ModerationListingService interface {
	ListPending(context.Context, users.VerifiedIdentity) ([]listings.Listing, error)
	Approve(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error)
	Reject(context.Context, users.VerifiedIdentity, uuid.UUID, int, string) (listings.Listing, error)
}
type moderationListingHandler struct{ service ModerationListingService }

func NewModerationListingHandler(service ModerationListingService) http.Handler {
	return moderationListingHandler{service: service}
}
func (h moderationListingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/moderation/listings")
	if suffix == "" || suffix == "/" {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, http.StatusText(405), 405)
			return
		}
		values, err := h.service.ListPending(r.Context(), identity)
		if err != nil {
			writeModerationError(w, err, id)
			return
		}
		out := make([]listingResponse, len(values))
		for i, v := range values {
			out[i] = listingResponseFrom(v)
		}
		writeJSON(w, 200, listingsResponse{Listings: out}, id)
		return
	}
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	if len(parts) != 2 {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	listingID, err := uuid.Parse(parts[0])
	if err != nil || r.Method != http.MethodPost {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	switch parts[1] {
	case "approve":
		revision, ok := decodeRevision(r.Body)
		if !ok {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		value, err := h.service.Approve(r.Context(), identity, listingID, revision)
		if err != nil {
			writeModerationError(w, err, id)
			return
		}
		writeJSON(w, 200, listingResponseFrom(value), id)
	case "reject":
		d := json.NewDecoder(io.LimitReader(r.Body, maxListingRequestBytes+1))
		d.DisallowUnknownFields()
		var value struct {
			Revision *int    `json:"revision"`
			Reason   *string `json:"reason"`
		}
		if err := d.Decode(&value); err != nil || value.Revision == nil || value.Reason == nil || *value.Revision < 1 {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		var extra any
		if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		listing, err := h.service.Reject(r.Context(), identity, listingID, *value.Revision, *value.Reason)
		if err != nil {
			writeModerationError(w, err, id)
			return
		}
		writeJSON(w, 200, listingResponseFrom(listing), id)
	default:
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	}
}
func writeModerationError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, moderation.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, moderation.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, moderation.ErrUnavailable):
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	default:
		writeListingError(w, err, id)
	}
}
