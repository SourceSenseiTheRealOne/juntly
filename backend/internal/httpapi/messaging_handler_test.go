package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/messaging"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestMessagingHandlerSendsParticipantMessage(t *testing.T) {
	t.Parallel()
	conversationID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	service := &recordingMessagingService{message: messaging.Message{ID: uuid.New(), ConversationID: conversationID, SenderID: uuid.New(), Body: "Boa tarde"}}
	handler := authn.RequireVerifiedIdentity(handlerVerifier{}, httpapi.NewMessagingHandler(service))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/conversations/"+conversationID.String()+"/messages", strings.NewReader(`{"body":"Boa tarde"}`))
	request.Header.Set("Authorization", "Bearer synthetic")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated || service.sentBody != "Boa tarde" || service.conversationID != conversationID {
		t.Fatalf("status/body/conversation = %d/%q/%s", response.Code, service.sentBody, service.conversationID)
	}
	if !strings.Contains(response.Body.String(), `"body":"Boa tarde"`) {
		t.Fatalf("response = %s", response.Body.String())
	}
}

func TestMessagingHandlerRejectsUnknownMessageFields(t *testing.T) {
	t.Parallel()
	service := &recordingMessagingService{}
	handler := authn.RequireVerifiedIdentity(handlerVerifier{}, httpapi.NewMessagingHandler(service))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/me/conversations/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/messages", strings.NewReader(`{"body":"hello","admin":true}`))
	request.Header.Set("Authorization", "Bearer synthetic")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || service.sentBody != "" {
		t.Fatalf("status/body = %d/%q", response.Code, service.sentBody)
	}
}

type handlerVerifier struct{}

func (handlerVerifier) Verify(context.Context, string) (users.VerifiedIdentity, error) {
	return users.VerifiedIdentity{Subject: "user_customer"}, nil
}

type recordingMessagingService struct {
	message        messaging.Message
	sentBody       string
	conversationID uuid.UUID
}

func (s *recordingMessagingService) Start(context.Context, users.VerifiedIdentity, uuid.UUID) (messaging.Conversation, error) {
	return messaging.Conversation{}, nil
}
func (s *recordingMessagingService) List(context.Context, users.VerifiedIdentity) ([]messaging.Conversation, error) {
	return nil, nil
}
func (s *recordingMessagingService) ListMessages(context.Context, users.VerifiedIdentity, uuid.UUID) ([]messaging.Message, error) {
	return nil, nil
}
func (s *recordingMessagingService) Send(_ context.Context, _ users.VerifiedIdentity, id uuid.UUID, body string) (messaging.Message, error) {
	s.conversationID, s.sentBody = id, body
	return s.message, nil
}
func (s *recordingMessagingService) SetBlocked(context.Context, users.VerifiedIdentity, uuid.UUID, bool) error {
	return nil
}
func (s *recordingMessagingService) Report(context.Context, users.VerifiedIdentity, uuid.UUID, *uuid.UUID, string) error {
	return nil
}
func (s *recordingMessagingService) Preferences(context.Context, users.VerifiedIdentity) (messaging.NotificationPreferences, error) {
	return messaging.NotificationPreferences{}, nil
}
func (s *recordingMessagingService) ReplacePreferences(context.Context, users.VerifiedIdentity, messaging.NotificationPreferences) (messaging.NotificationPreferences, error) {
	return messaging.NotificationPreferences{}, nil
}
func (s *recordingMessagingService) Notifications(context.Context, users.VerifiedIdentity) ([]messaging.Notification, error) {
	return nil, nil
}
func (s *recordingMessagingService) MarkNotificationRead(context.Context, users.VerifiedIdentity, uuid.UUID) error {
	return nil
}
