package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

const maxContactChannelRequestBytes = 2048

type ContactChannelService interface {
	Get(context.Context, users.VerifiedIdentity) ([]contactreveal.ChannelStatus, error)
	Put(context.Context, users.VerifiedIdentity, contactreveal.ReplaceChannel) (contactreveal.ChannelStatus, error)
}

type contactChannelHandler struct{ service ContactChannelService }

type channelStatusResponse struct {
	Channel       contactreveal.Channel `json:"channel"`
	Configured    bool                  `json:"configured"`
	Enabled       bool                  `json:"enabled"`
	RevealConsent bool                  `json:"revealConsent"`
}

type contactChannelsResponse struct {
	Channels []channelStatusResponse `json:"channels"`
}
type replaceContactChannelRequest struct {
	Channel       *contactreveal.Channel `json:"channel"`
	Contact       *string                `json:"contact"`
	Enabled       *bool                  `json:"enabled"`
	RevealConsent *bool                  `json:"revealConsent"`
}

func NewContactChannelHandler(service ContactChannelService) http.Handler {
	return contactChannelHandler{service: service}
}

func (h contactChannelHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
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
	if request.Method == http.MethodGet {
		values, err := h.service.Get(request.Context(), identity)
		if err != nil {
			writeContactRevealError(w, err, id)
			return
		}
		out := make([]channelStatusResponse, len(values))
		for index, value := range values {
			out[index] = statusResponse(value)
		}
		writeJSON(w, http.StatusOK, contactChannelsResponse{Channels: out}, id)
		return
	}
	input, valid := decodeContactChannel(request.Body)
	if !valid {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", id)
		return
	}
	value, err := h.service.Put(request.Context(), identity, input)
	if err != nil {
		writeContactRevealError(w, err, id)
		return
	}
	writeJSON(w, http.StatusOK, statusResponse(value), id)
}

func decodeContactChannel(body io.Reader) (contactreveal.ReplaceChannel, bool) {
	decoder := json.NewDecoder(io.LimitReader(body, maxContactChannelRequestBytes+1))
	decoder.DisallowUnknownFields()
	var value replaceContactChannelRequest
	if err := decoder.Decode(&value); err != nil || value.Channel == nil || value.Contact == nil || value.Enabled == nil || value.RevealConsent == nil {
		return contactreveal.ReplaceChannel{}, false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return contactreveal.ReplaceChannel{}, false
	}
	return contactreveal.ReplaceChannel{Channel: *value.Channel, Contact: *value.Contact, Enabled: *value.Enabled, RevealConsent: *value.RevealConsent}, true
}

func statusResponse(value contactreveal.ChannelStatus) channelStatusResponse {
	return channelStatusResponse{Channel: value.Channel, Configured: value.Configured, Enabled: value.Enabled, RevealConsent: value.RevealConsent}
}

func writeContactRevealError(w http.ResponseWriter, err error, id string) {
	switch {
	case errors.Is(err, contactreveal.ErrUnauthorized):
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", id)
	case errors.Is(err, contactreveal.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden", id)
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", id)
	}
}
