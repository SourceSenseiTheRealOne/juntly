package provideraccess

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/accounts"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestRequireProviderRejectsInvalidIdentity(t *testing.T) {
	t.Parallel()

	identities := &recordingIdentityService{err: users.ErrInvalidIdentity}
	capabilities := &recordingCapabilityReader{}
	_, err := NewService(identities, capabilities).RequireProvider(context.Background(), users.VerifiedIdentity{})

	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("error = %v, want ErrUnauthorized", err)
	}
	if capabilities.calls != 0 {
		t.Fatalf("capability calls = %d, want 0", capabilities.calls)
	}
}

func TestRequireProviderRejectsDisabledCapability(t *testing.T) {
	t.Parallel()

	identities := &recordingIdentityService{user: testInternalUser()}
	capabilities := &recordingCapabilityReader{account: accounts.Account{CustomerEnabled: true}}
	_, err := NewService(identities, capabilities).RequireProvider(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_provider"},
	)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestRequireProviderReturnsVerifiedInternalOwner(t *testing.T) {
	t.Parallel()

	owner := testInternalUser()
	identities := &recordingIdentityService{user: owner}
	capabilities := &recordingCapabilityReader{account: accounts.Account{CustomerEnabled: true, ProviderEnabled: true}}

	result, err := NewService(identities, capabilities).RequireProvider(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_provider"},
	)

	if err != nil || result != owner {
		t.Fatalf("owner = %#v, err = %v, want %#v", result, err, owner)
	}
	if identities.calls != 1 || capabilities.calls != 1 {
		t.Fatalf("calls = identity:%d capability:%d, want 1:1", identities.calls, capabilities.calls)
	}
}

func TestRequireProviderMapsDependencyFailure(t *testing.T) {
	t.Parallel()

	identities := &recordingIdentityService{user: testInternalUser()}
	capabilities := &recordingCapabilityReader{err: accounts.ErrUnavailable}
	_, err := NewService(identities, capabilities).RequireProvider(
		context.Background(),
		users.VerifiedIdentity{Subject: "user_provider"},
	)

	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error = %v, want ErrUnavailable", err)
	}
}

func testInternalUser() users.InternalUser {
	return users.InternalUser{
		ID:        uuid.MustParse("b00f5bf7-72a8-4a7d-bb23-2ef4c4daf3fb"),
		CreatedAt: time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC),
	}
}

type recordingIdentityService struct {
	user  users.InternalUser
	err   error
	calls int
}

func (s *recordingIdentityService) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	s.calls++
	return s.user, false, s.err
}

type recordingCapabilityReader struct {
	account accounts.Account
	err     error
	calls   int
}

func (r *recordingCapabilityReader) Get(context.Context, users.VerifiedIdentity) (accounts.Account, error) {
	r.calls++
	return r.account, r.err
}
