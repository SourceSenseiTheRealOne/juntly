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
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/messaging"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

const maxMessagingRequestBytes = 8 * 1024

type MessagingService interface {
	Start(context.Context, users.VerifiedIdentity, uuid.UUID) (messaging.Conversation, error)
	List(context.Context, users.VerifiedIdentity) ([]messaging.Conversation, error)
	ListMessages(context.Context, users.VerifiedIdentity, uuid.UUID) ([]messaging.Message, error)
	Send(context.Context, users.VerifiedIdentity, uuid.UUID, string) (messaging.Message, error)
	SetBlocked(context.Context, users.VerifiedIdentity, uuid.UUID, bool) error
	Report(context.Context, users.VerifiedIdentity, uuid.UUID, *uuid.UUID, string) error
	Preferences(context.Context, users.VerifiedIdentity) (messaging.NotificationPreferences, error)
	ReplacePreferences(context.Context, users.VerifiedIdentity, messaging.NotificationPreferences) (messaging.NotificationPreferences, error)
	Notifications(context.Context, users.VerifiedIdentity) ([]messaging.Notification, error)
	MarkNotificationRead(context.Context, users.VerifiedIdentity, uuid.UUID) error
}

type messagingHandler struct{ service MessagingService }

func NewMessagingHandler(service MessagingService) http.Handler {
	return messagingHandler{service: service}
}

type conversationResponse struct {
	ID         uuid.UUID  `json:"id"`
	ListingID  *uuid.UUID `json:"listingId"`
	CustomerID uuid.UUID  `json:"customerId"`
	ProviderID uuid.UUID  `json:"providerId"`
	Blocked    bool       `json:"blocked"`
	CreatedAt  string     `json:"createdAt"`
	UpdatedAt  string     `json:"updatedAt"`
}
type messageResponse struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversationId"`
	SenderID       uuid.UUID `json:"senderId"`
	Body           string    `json:"body"`
	CreatedAt      string    `json:"createdAt"`
}
type notificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Kind      string    `json:"kind"`
	Read      bool      `json:"read"`
	CreatedAt string    `json:"createdAt"`
}
type startConversationRequest struct {
	ListingID *uuid.UUID `json:"listingId"`
}
type sendMessageRequest struct {
	Body *string `json:"body"`
}
type blockConversationRequest struct {
	Blocked *bool `json:"blocked"`
}
type reportConversationRequest struct {
	MessageID *uuid.UUID `json:"messageId"`
	Reason    *string    `json:"reason"`
}
type notificationPreferencesRequest struct {
	InAppEnabled *bool `json:"inAppEnabled"`
	EmailEnabled *bool `json:"emailEnabled"`
}
type notificationPreferencesResponse struct {
	InAppEnabled bool `json:"inAppEnabled"`
	EmailEnabled bool `json:"emailEnabled"`
}

func (h messagingHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	requestID := requestIDFromHeader(request.Header.Get(RequestIDHeader))
	identity, ok := authn.IdentityFromContext(request.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
		return
	}
	if h.service == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
		return
	}
	path := strings.TrimSuffix(request.URL.Path, "/")
	if strings.HasPrefix(path, "/api/v1/me/conversations") {
		h.conversations(w, request, identity, requestID, path)
		return
	}
	if strings.HasPrefix(path, "/api/v1/me/notifications") {
		h.notifications(w, request, identity, requestID, path)
		return
	}
	writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Not found", requestID)
}

func (h messagingHandler) conversations(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, requestID, path string) {
	base := "/api/v1/me/conversations"
	if path == base {
		switch r.Method {
		case http.MethodGet:
			values, err := h.service.List(r.Context(), identity)
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			responses := make([]conversationResponse, 0, len(values))
			for _, v := range values {
				responses = append(responses, conversationJSON(v))
			}
			writeJSON(w, http.StatusOK, map[string]any{"conversations": responses}, requestID)
		case http.MethodPost:
			var input startConversationRequest
			if !decodeMessaging(r.Body, &input) || input.ListingID == nil {
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			v, err := h.service.Start(r.Context(), identity, *input.ListingID)
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusCreated, conversationJSON(v), requestID)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, base+"/"), "/")
	if len(parts) != 2 {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Not found", requestID)
		return
	}
	conversationID, err := uuid.Parse(parts[0])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
		return
	}
	switch parts[1] {
	case "messages":
		if r.Method == http.MethodGet {
			values, err := h.service.ListMessages(r.Context(), identity, conversationID)
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			responses := make([]messageResponse, 0, len(values))
			for _, v := range values {
				responses = append(responses, messageJSON(v))
			}
			writeJSON(w, http.StatusOK, map[string]any{"messages": responses}, requestID)
			return
		}
		if r.Method == http.MethodPost {
			var input sendMessageRequest
			if !decodeMessaging(r.Body, &input) || input.Body == nil {
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			v, err := h.service.Send(r.Context(), identity, conversationID, *input.Body)
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusCreated, messageJSON(v), requestID)
			return
		}
	case "block":
		if r.Method == http.MethodPut {
			var input blockConversationRequest
			if !decodeMessaging(r.Body, &input) || input.Blocked == nil {
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			if err := h.service.SetBlocked(r.Context(), identity, conversationID, *input.Blocked); err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"blocked": *input.Blocked}, requestID)
			return
		}
	case "reports":
		if r.Method == http.MethodPost {
			var input reportConversationRequest
			if !decodeMessaging(r.Body, &input) || input.Reason == nil {
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			if err := h.service.Report(r.Context(), identity, conversationID, input.MessageID, *input.Reason); err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]bool{"reported": true}, requestID)
			return
		}
	}
	w.Header().Set("Allow", "GET, POST, PUT")
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (h messagingHandler) notifications(w http.ResponseWriter, r *http.Request, identity users.VerifiedIdentity, requestID, path string) {
	base := "/api/v1/me/notifications"
	if path == base {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		values, err := h.service.Notifications(r.Context(), identity)
		if err != nil {
			writeMessagingError(w, err, requestID)
			return
		}
		responses := make([]notificationResponse, 0, len(values))
		for _, v := range values {
			responses = append(responses, notificationJSON(v))
		}
		writeJSON(w, http.StatusOK, map[string]any{"notifications": responses}, requestID)
		return
	}
	if path == base+"/preferences" {
		switch r.Method {
		case http.MethodGet:
			v, err := h.service.Preferences(r.Context(), identity)
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusOK, preferencesJSON(v), requestID)
		case http.MethodPut:
			var input notificationPreferencesRequest
			if !decodeMessaging(r.Body, &input) || input.InAppEnabled == nil || input.EmailEnabled == nil {
				writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
				return
			}
			v, err := h.service.ReplacePreferences(r.Context(), identity, messaging.NotificationPreferences{InAppEnabled: *input.InAppEnabled, EmailEnabled: *input.EmailEnabled})
			if err != nil {
				writeMessagingError(w, err, requestID)
				return
			}
			writeJSON(w, http.StatusOK, preferencesJSON(v), requestID)
		default:
			w.Header().Set("Allow", "GET, PUT")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		return
	}
	parts := strings.Split(strings.TrimPrefix(path, base+"/"), "/")
	if len(parts) == 2 && parts[1] == "read" && r.Method == http.MethodPost {
		id, err := uuid.Parse(parts[0])
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
			return
		}
		if err = h.service.MarkNotificationRead(r.Context(), identity, id); err != nil {
			writeMessagingError(w, err, requestID)
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"read": true}, requestID)
		return
	}
	writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Not found", requestID)
}

func decodeMessaging(body io.Reader, target any) bool {
	decoder := json.NewDecoder(io.LimitReader(body, maxMessagingRequestBytes+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}
func conversationJSON(v messaging.Conversation) conversationResponse {
	return conversationResponse{ID: v.ID, ListingID: v.ListingID, CustomerID: v.CustomerID, ProviderID: v.ProviderID, Blocked: v.BlockedBy != nil, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano), UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339Nano)}
}
func messageJSON(v messaging.Message) messageResponse {
	return messageResponse{ID: v.ID, ConversationID: v.ConversationID, SenderID: v.SenderID, Body: v.Body, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
func notificationJSON(v messaging.Notification) notificationResponse {
	return notificationResponse{ID: v.ID, Kind: v.Kind, Read: v.ReadAt != nil, CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
func preferencesJSON(v messaging.NotificationPreferences) notificationPreferencesResponse {
	return notificationPreferencesResponse{InAppEnabled: v.InAppEnabled, EmailEnabled: v.EmailEnabled}
}
func writeMessagingError(w http.ResponseWriter, err error, requestID string) {
	switch {
	case errors.Is(err, messaging.ErrInvalid):
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request", requestID)
	case errors.Is(err, messaging.ErrUnauthorized):
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", requestID)
	case errors.Is(err, messaging.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Forbidden", requestID)
	case errors.Is(err, messaging.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Not found", requestID)
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service unavailable", requestID)
	}
}
