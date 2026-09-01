package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/reviews"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

const maxReviewRequestBytes = 4 * 1024

type ReviewService interface {
	Create(context.Context, users.VerifiedIdentity, reviews.CreateReview) (reviews.Review, error)
	ListForProvider(context.Context, users.VerifiedIdentity) ([]reviews.Review, error)
	Respond(context.Context, users.VerifiedIdentity, uuid.UUID, string) (reviews.Review, error)
	Aggregate(context.Context, uuid.UUID) (reviews.Aggregate, error)
}
type reviewHandler struct{ service ReviewService }

func NewReviewHandler(s ReviewService) http.Handler { return reviewHandler{service: s} }

type createReviewRequest struct {
	BookingID *uuid.UUID `json:"bookingId"`
	Rating    *int       `json:"rating"`
	Body      *string    `json:"body"`
}
type responseRequest struct {
	Response *string `json:"response"`
}

func (h reviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := requestIDFromHeader(r.Header.Get(RequestIDHeader))
	if h.service == nil {
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/v1/public/providers/") {
		parts := strings.Split(strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/api/v1/public/providers/"), "/")
		if len(parts) != 2 || parts[1] != "rating" || r.Method != http.MethodGet {
			writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
			return
		}
		provider, err := uuid.Parse(parts[0])
		if err != nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		v, err := h.service.Aggregate(r.Context(), provider)
		if err != nil {
			writeReviewError(w, err, id)
			return
		}
		writeJSON(w, 200, map[string]any{"providerId": v.ProviderID, "averageRating": v.AverageRating, "reviewCount": v.ReviewCount}, id)
		return
	}
	identity, ok := authn.IdentityFromContext(r.Context())
	if !ok {
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "/api/v1/me/reviews" {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), 405)
			return
		}
		var body createReviewRequest
		if !decodeReview(r.Body, &body) || body.BookingID == nil || body.Rating == nil || body.Body == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		v, err := h.service.Create(r.Context(), identity, reviews.CreateReview{BookingID: *body.BookingID, Rating: *body.Rating, Body: *body.Body})
		if err != nil {
			writeReviewError(w, err, id)
			return
		}
		writeJSON(w, 201, reviewJSON(v), id)
		return
	}
	if path == "/api/v1/me/reviews/provider" && r.Method == http.MethodGet {
		values, err := h.service.ListForProvider(r.Context(), identity)
		if err != nil {
			writeReviewError(w, err, id)
			return
		}
		out := make([]map[string]any, 0, len(values))
		for _, v := range values {
			out = append(out, reviewJSON(v))
		}
		writeJSON(w, 200, map[string]any{"reviews": out}, id)
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/me/reviews/"), "/")
	if len(parts) == 2 && parts[1] == "response" && r.Method == http.MethodPut {
		reviewID, err := uuid.Parse(parts[0])
		if err != nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		var body responseRequest
		if !decodeReview(r.Body, &body) || body.Response == nil {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		v, err := h.service.Respond(r.Context(), identity, reviewID, *body.Response)
		if err != nil {
			writeReviewError(w, err, id)
			return
		}
		writeJSON(w, 200, reviewJSON(v), id)
		return
	}
	writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
}
func decodeReview(r io.Reader, target any) bool {
	d := json.NewDecoder(io.LimitReader(r, maxReviewRequestBytes+1))
	d.DisallowUnknownFields()
	if d.Decode(target) != nil {
		return false
	}
	var extra any
	return errors.Is(d.Decode(&extra), io.EOF)
}
func reviewJSON(v reviews.Review) map[string]any {
	return map[string]any{"id": v.ID, "bookingId": v.BookingID, "customerId": v.CustomerID, "providerId": v.ProviderID, "rating": v.Rating, "body": v.Body, "providerResponse": v.ProviderResponse, "verifiedBooking": v.VerifiedBooking, "state": v.State, "createdAt": v.CreatedAt, "updatedAt": v.UpdatedAt}
}
func writeReviewError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, reviews.ErrInvalid):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, reviews.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, reviews.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, reviews.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	case errors.Is(err, reviews.ErrNotFound):
		writeAPIError(w, 404, "NOT_FOUND", "Not found", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
