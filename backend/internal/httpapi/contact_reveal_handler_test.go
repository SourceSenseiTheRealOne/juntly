package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/httpapi"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestContactRevealHandlerStrictlyRevealsOnlyToVerifiedCustomer(t *testing.T) {
	t.Parallel()
	listingID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	service := &recordingContactRevealService{result: contactreveal.RevealedContact{Channel: contactreveal.ChannelPhone, Value: "revealed-contact"}}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "customer"}}, httpapi.NewContactRevealHandler(service))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals", strings.NewReader(`{"channel":"phone"}`))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_contact_reveal")
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.calls != 1 || service.listingID != listingID || service.channel != contactreveal.ChannelPhone {
		t.Fatalf("status/service = %d/%#v", response.Code, service)
	}
	if !strings.Contains(response.Body.String(), "revealed-contact") || strings.Contains(response.Body.String(), "internalUserId") {
		t.Fatalf("unexpected reveal response: %s", response.Body.String())
	}
	for _, body := range []string{`{}`, `{"channel":"phone","extra":true}`, `{"channel":"email"}`} {
		service := &recordingContactRevealService{}
		handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "customer"}}, httpapi.NewContactRevealHandler(service))
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals", strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer synthetic-token")
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest || service.calls != 0 {
			t.Fatalf("body/status/calls = %s/%d/%d", body, response.Code, service.calls)
		}
	}
}

func TestContactRevealHandlerReturnsGenericForbiddenWithoutContact(t *testing.T) {
	t.Parallel()
	service := &recordingContactRevealService{err: contactreveal.ErrForbidden}
	handler := authn.RequireVerifiedIdentity(staticVerifier{identity: users.VerifiedIdentity{Subject: "customer"}}, httpapi.NewContactRevealHandler(service))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/listings/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa/contact-reveals", strings.NewReader(`{"channel":"phone"}`))
	request.Header.Set("Authorization", "Bearer synthetic-token")
	request.Header.Set(httpapi.RequestIDHeader, "req_contact_forbidden")
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d", response.Code)
	}
	assertErrorResponse(t, response, "FORBIDDEN", "Forbidden", "req_contact_forbidden")
}

type recordingContactRevealService struct {
	calls     int
	listingID uuid.UUID
	channel   contactreveal.Channel
	result    contactreveal.RevealedContact
	err       error
}

func (s *recordingContactRevealService) Reveal(_ context.Context, _ users.VerifiedIdentity, listingID uuid.UUID, channel contactreveal.Channel) (contactreveal.RevealedContact, error) {
	s.calls++
	s.listingID = listingID
	s.channel = channel
	return s.result, s.err
}

var _ = errors.Is
