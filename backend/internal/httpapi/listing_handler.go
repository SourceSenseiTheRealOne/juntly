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
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/listings"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxListingRequestBytes = 16 * 1024

type ListingService interface {
	Create(context.Context, users.VerifiedIdentity, listings.CreateListing) (listings.Listing, error)
	ReplaceDraft(context.Context, users.VerifiedIdentity, uuid.UUID, int, listings.CreateListing) (listings.Listing, error)
	Get(context.Context, users.VerifiedIdentity, uuid.UUID) (*listings.Listing, error)
	List(context.Context, users.VerifiedIdentity) ([]listings.Listing, error)
	Submit(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error)
	Pause(context.Context, users.VerifiedIdentity, uuid.UUID, int) (listings.Listing, error)
	Archive(context.Context, users.VerifiedIdentity, uuid.UUID, listings.State, int) (listings.Listing, error)
	CreateUploadIntent(context.Context, users.VerifiedIdentity, uuid.UUID, listingmedia.UploadRequest) (listingmedia.UploadIntent, error)
}
type listingHandler struct{ service ListingService }
type listingRequest struct {
	CategoryID        *uuid.UUID `json:"categoryId"`
	PrimaryLocalityID *uuid.UUID `json:"primaryLocalityId"`
	Title             *string    `json:"title"`
	Description       *string    `json:"description"`
	PriceType         *string    `json:"priceType"`
	PriceMinor        **int      `json:"priceMinor"`
	Currency          *string    `json:"currency"`
	TravelsToCustomer *bool      `json:"travelsToCustomer"`
	ReceivesCustomer  *bool      `json:"receivesCustomer"`
	RemoteServices    *bool      `json:"remoteServices"`
	Revision          *int       `json:"revision"`
}
type listingResponse struct {
	ID                uuid.UUID          `json:"id"`
	CategoryID        uuid.UUID          `json:"categoryId"`
	PrimaryLocalityID uuid.UUID          `json:"primaryLocalityId"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	PriceType         listings.PriceType `json:"priceType"`
	PriceMinor        *int               `json:"priceMinor"`
	Currency          string             `json:"currency"`
	TravelsToCustomer bool               `json:"travelsToCustomer"`
	ReceivesCustomer  bool               `json:"receivesCustomer"`
	RemoteServices    bool               `json:"remoteServices"`
	State             listings.State     `json:"state"`
	Revision          int                `json:"revision"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
}
type listingsResponse struct {
	Listings []listingResponse `json:"listings"`
}
type uploadCapabilityResponse struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}
type uploadIntentResponse struct {
	MediaID    uuid.UUID                `json:"mediaId"`
	Capability uploadCapabilityResponse `json:"capability"`
}

func NewListingHandler(service ListingService) http.Handler { return listingHandler{service: service} }
func (h listingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/me/listings")
	if suffix == "" || suffix == "/" {
		h.collection(w, r, identity, id)
		return
	}
	parts := strings.Split(strings.Trim(suffix, "/"), "/")
	listingID, err := uuid.Parse(parts[0])
	if err != nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	if len(parts) == 2 {
		switch parts[1] {
		case "submit":
			h.submit(w, r, identity, listingID, id)
		case "pause":
			h.pause(w, r, identity, listingID, id)
		case "archive":
			h.archive(w, r, identity, listingID, id)
		default:
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		}
		return
	}
	if len(parts) == 3 && parts[1] == "media" && parts[2] == "upload-intents" {
		h.uploadIntent(w, r, identity, listingID, id)
		return
	}
	if len(parts) != 1 {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	h.item(w, r, identity, listingID, id)
}
func (h listingHandler) collection(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, id string) {
	switch r.Method {
	case http.MethodGet:
		values, err := h.service.List(r.Context(), identity)
		if err != nil {
			writeListingError(w, err, id)
			return
		}
		out := make([]listingResponse, len(values))
		for i, v := range values {
			out[i] = listingResponseFrom(v)
		}
		writeJSON(w, 200, listingsResponse{Listings: out}, id)
	case http.MethodPost:
		input, _, ok := decodeListing(r.Body, false)
		if !ok {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		value, err := h.service.Create(r.Context(), identity, input)
		if err != nil {
			writeListingError(w, err, id)
			return
		}
		writeJSON(w, 200, listingResponseFrom(value), id)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}
func (h listingHandler) item(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, listingID uuid.UUID, id string) {
	switch r.Method {
	case http.MethodGet:
		value, err := h.service.Get(r.Context(), identity, listingID)
		if err != nil {
			writeListingError(w, err, id)
			return
		}
		if value == nil {
			writeListingError(w, listings.ErrConflict, id)
			return
		}
		writeJSON(w, 200, listingResponseFrom(*value), id)
	case http.MethodPut:
		input, revision, ok := decodeListing(r.Body, true)
		if !ok {
			writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
			return
		}
		value, err := h.service.ReplaceDraft(r.Context(), identity, listingID, revision, input)
		if err != nil {
			writeListingError(w, err, id)
			return
		}
		writeJSON(w, 200, listingResponseFrom(value), id)
	default:
		w.Header().Set("Allow", "GET, PUT")
		http.Error(w, http.StatusText(405), 405)
	}
}
func (h listingHandler) submit(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, listingID uuid.UUID, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	revision, ok := decodeRevision(r.Body)
	if !ok {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	value, err := h.service.Submit(r.Context(), identity, listingID, revision)
	if err != nil {
		writeListingError(w, err, id)
		return
	}
	writeJSON(w, 200, listingResponseFrom(value), id)
}
func (h listingHandler) pause(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, listingID uuid.UUID, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	revision, ok := decodeRevision(r.Body)
	if !ok {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	value, err := h.service.Pause(r.Context(), identity, listingID, revision)
	if err != nil {
		writeListingError(w, err, id)
		return
	}
	writeJSON(w, 200, listingResponseFrom(value), id)
}
func (h listingHandler) archive(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, listingID uuid.UUID, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	d := json.NewDecoder(io.LimitReader(r.Body, maxListingRequestBytes+1))
	d.DisallowUnknownFields()
	var value struct {
		Revision *int    `json:"revision"`
		State    *string `json:"state"`
	}
	if err := d.Decode(&value); err != nil || value.Revision == nil || value.State == nil || *value.Revision < 1 {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	from := listings.State(*value.State)
	if from != listings.StateDraft && from != listings.StateRejected && from != listings.StateActive && from != listings.StatePaused {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	listing, err := h.service.Archive(r.Context(), identity, listingID, from, *value.Revision)
	if err != nil {
		writeListingError(w, err, id)
		return
	}
	writeJSON(w, 200, listingResponseFrom(listing), id)
}
func (h listingHandler) uploadIntent(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, listingID uuid.UUID, id string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	d := json.NewDecoder(io.LimitReader(r.Body, maxListingRequestBytes+1))
	d.DisallowUnknownFields()
	var value struct {
		Ordinal        *int    `json:"ordinal"`
		ContentType    *string `json:"contentType"`
		ByteSize       *int64  `json:"byteSize"`
		ChecksumSHA256 *string `json:"checksumSha256"`
	}
	if err := d.Decode(&value); err != nil || value.Ordinal == nil || value.ContentType == nil || value.ByteSize == nil || value.ChecksumSHA256 == nil {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	intent, err := h.service.CreateUploadIntent(r.Context(), identity, listingID, listingmedia.UploadRequest{Ordinal: *value.Ordinal, ContentType: *value.ContentType, ByteSize: *value.ByteSize, ChecksumSHA256: *value.ChecksumSHA256})
	if err != nil {
		writeListingError(w, err, id)
		return
	}
	writeJSON(w, 200, uploadIntentResponse{MediaID: intent.MediaID, Capability: uploadCapabilityResponse{URL: intent.Capability.URL, Method: intent.Capability.Method, Headers: intent.Capability.Headers}}, id)
}
func decodeListing(body io.Reader, revisionRequired bool) (listings.CreateListing, int, bool) {
	d := json.NewDecoder(io.LimitReader(body, maxListingRequestBytes+1))
	d.DisallowUnknownFields()
	var v listingRequest
	if err := d.Decode(&v); err != nil {
		return listings.CreateListing{}, 0, false
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return listings.CreateListing{}, 0, false
	}
	if v.CategoryID == nil || v.PrimaryLocalityID == nil || v.Title == nil || v.Description == nil || v.PriceType == nil || v.PriceMinor == nil || v.Currency == nil || v.TravelsToCustomer == nil || v.ReceivesCustomer == nil || v.RemoteServices == nil || (revisionRequired && v.Revision == nil) {
		return listings.CreateListing{}, 0, false
	}
	var price *int
	if *v.PriceMinor != nil {
		value := **v.PriceMinor
		price = &value
	}
	return listings.CreateListing{CategoryID: *v.CategoryID, PrimaryLocalityID: *v.PrimaryLocalityID, Title: *v.Title, Description: *v.Description, PriceType: listings.PriceType(*v.PriceType), PriceMinor: price, Currency: *v.Currency, TravelsToCustomer: *v.TravelsToCustomer, ReceivesCustomer: *v.ReceivesCustomer, RemoteServices: *v.RemoteServices}, valueOrZero(v.Revision), true
}
func decodeRevision(body io.Reader) (int, bool) {
	d := json.NewDecoder(io.LimitReader(body, maxListingRequestBytes+1))
	d.DisallowUnknownFields()
	var v struct {
		Revision *int `json:"revision"`
	}
	if err := d.Decode(&v); err != nil || v.Revision == nil {
		return 0, false
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return 0, false
	}
	if *v.Revision < 1 {
		return 0, false
	}
	return *v.Revision, true
}
func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
func listingResponseFrom(v listings.Listing) listingResponse {
	return listingResponse{ID: v.ID, CategoryID: v.CategoryID, PrimaryLocalityID: v.PrimaryLocalityID, Title: v.Title, Description: v.Description, PriceType: v.PriceType, PriceMinor: v.PriceMinor, Currency: v.Currency, TravelsToCustomer: v.TravelsToCustomer, ReceivesCustomer: v.ReceivesCustomer, RemoteServices: v.RemoteServices, State: v.State, Revision: v.Revision, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func writeListingError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, provideraccess.ErrUnauthorized):
		writeAPIError(w, 401, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, provideraccess.ErrForbidden):
		writeAPIError(w, 403, "FORBIDDEN", "Forbidden", id)
	case errors.Is(err, listings.ErrInvalidListing):
		writeAPIError(w, 400, "INVALID_REQUEST", "Invalid request", id)
	case errors.Is(err, listings.ErrConflict):
		writeAPIError(w, 409, "CONFLICT", "Conflict", id)
	default:
		writeAPIError(w, 503, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
