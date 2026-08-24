package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxContactRevealRequestBytes = 1024

type ContactRevealService interface {
	Reveal(context.Context, users.VerifiedIdentity, uuid.UUID, contactreveal.Channel) (contactreveal.RevealedContact, error)
}

type contactRevealHandler struct{ service ContactRevealService }
type contactRevealRequest struct {
	Channel *contactreveal.Channel `json:"channel"`
}
type contactRevealResponse struct {
	Channel contactreveal.Channel `json:"channel"`
	Contact string                `json:"contact"`
}

func NewContactRevealHandler(service ContactRevealService) http.Handler {
	return contactRevealHandler{service: service}
}

func (h contactRevealHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	id := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(request.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", id)
		return
	}
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", id)
		return
	}
	listingID, valid := revealListingID(request.URL.Path)
	channel, validBody := decodeRevealChannel(request.Body)
	if !valid || !validBody {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	value, err := h.service.Reveal(request.Context(), identity, listingID, channel)
	if err != nil {
		writeContactRevealError(w, err, id)
		return
	}
	writeJSON(w, http.StatusOK, contactRevealResponse{Channel: value.Channel, Contact: value.Value}, id)
}

func revealListingID(path string) (uuid.UUID, bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/listings/"), "/")
	if len(parts) != 2 || parts[1] != "contact-reveals" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(parts[0])
	return id, err == nil
}

func decodeRevealChannel(body io.Reader) (contactreveal.Channel, bool) {
	decoder := json.NewDecoder(io.LimitReader(body, maxContactRevealRequestBytes+1))
	decoder.DisallowUnknownFields()
	var value contactRevealRequest
	if err := decoder.Decode(&value); err != nil || value.Channel == nil || (*value.Channel != contactreveal.ChannelPhone && *value.Channel != contactreveal.ChannelWhatsApp) {
		return "", false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return "", false
	}
	return *value.Channel, true
}
