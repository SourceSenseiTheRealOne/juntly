package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

func TestContactChannelHandlerReturnsStatusOnlyAndStrictlyConfiguresOneChannel(t *testing.T) {
	t.Parallel()
	service := &recordingContactChannelService{statuses: []contactreveal.ChannelStatus{{Channel: contactreveal.ChannelPhone, Configured: true, Enabled: true, RevealConsent: true}}, status: contactreveal.ChannelStatus{Channel: contactreveal.ChannelWhatsApp, Configured: true, Enabled: true, RevealConsent: true}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "provider"}}, httpapi.NewContactChannelHandler(service))
	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/me/contact-channels", nil)
	getRequest.Header.Set("Authorization", "Bearer synthetic-token")
	getRequest.Header.Set(httpapi.RequestIDHeader, "req_contact_status")
	handler.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK || service.getCalls != 1 {
		t.Fatalf("get status/calls = %d/%d", getResponse.Code, service.getCalls)
	}
	serialized := getResponse.Body.String()
	for _, prohibited := range []string{"ciphertext", "nonce", "keyVersion", "contact", "synthetic-token"} {
		if strings.Contains(strings.ToLower(serialized), strings.ToLower(prohibited)) {
			t.Fatalf("status response leaks %q: %s", prohibited, serialized)
		}
	}
	putResponse := httptest.NewRecorder()
	putRequest := httptest.NewRequest(http.MethodPut, "/api/v1/me/contact-channels", strings.NewReader(`{"channel":"whatsapp","contact":"+12025550123","enabled":true,"revealConsent":true}`))
	putRequest.Header.Set("Authorization", "Bearer synthetic-token")
	putRequest.Header.Set(httpapi.RequestIDHeader, "req_contact_put")
	handler.ServeHTTP(putResponse, putRequest)
	if putResponse.Code != http.StatusOK || service.putCalls != 1 || service.input.Contact == "" {
		t.Fatalf("put status/calls/input = %d/%d/%#v", putResponse.Code, service.putCalls, service.input)
	}
	var body map[string]any
	if err := json.NewDecoder(putResponse.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, exists := body["contact"]; exists {
		t.Fatalf("put response leaks contact: %#v", body)
	}
}

type recordingContactChannelService struct {
	statuses []contactreveal.ChannelStatus
	status   contactreveal.ChannelStatus
	input    contactreveal.ReplaceChannel
	getCalls int
	putCalls int
}

func (s *recordingContactChannelService) Get(context.Context, users.VerifiedIdentity) ([]contactreveal.ChannelStatus, error) {
	s.getCalls++
	return s.statuses, nil
}
func (s *recordingContactChannelService) Put(_ context.Context, _ users.VerifiedIdentity, input contactreveal.ReplaceChannel) (contactreveal.ChannelStatus, error) {
	s.putCalls++
	s.input = input
	return s.status, nil
}
