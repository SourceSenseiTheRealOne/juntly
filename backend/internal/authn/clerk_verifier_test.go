package authn

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"
	"time"
)

func TestNewClerkVerifierRejectsMissingVerificationMaterial(t *testing.T) {
	t.Parallel()

	_, err := NewClerkVerifier(ClerkVerifierConfig{
		AuthorizedParties: []string{"http://localhost:4200"},
	})

	if !errors.Is(err, ErrInvalidClerkVerifierConfig) {
		t.Fatalf("error = %v, want ErrInvalidClerkVerifierConfig", err)
	}
}

func TestNewClerkVerifierRejectsMissingAuthorizedParties(t *testing.T) {
	t.Parallel()

	_, err := NewClerkVerifier(ClerkVerifierConfig{JWTKey: "synthetic-public-key"})

	if !errors.Is(err, ErrInvalidClerkVerifierConfig) {
		t.Fatalf("error = %v, want ErrInvalidClerkVerifierConfig", err)
	}
}

func TestClerkVerifierRejectsMalformedTokenWithoutIdentity(t *testing.T) {
	t.Parallel()

	verifier, err := NewClerkVerifier(ClerkVerifierConfig{
		JWTKey:            testPublicKey(t),
		AuthorizedParties: []string{"http://localhost:4200"},
	})
	if err != nil {
		t.Fatalf("new Clerk verifier: %v", err)
	}

	identity, err := verifier.Verify(context.Background(), "not-a-jwt")

	if err == nil {
		t.Fatal("Verify error = nil, want rejection")
	}
	if identity.Subject != "" {
		t.Fatalf("identity subject = %q, want empty", identity.Subject)
	}
}

func testPublicKey(t *testing.T) string {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func TestNewClerkVerifierAcceptsBoundedClockSkew(t *testing.T) {
	t.Parallel()

	verifier, err := NewClerkVerifier(ClerkVerifierConfig{
		JWTKey:            testPublicKey(t),
		AuthorizedParties: []string{"http://localhost:4200"},
		ClockSkew:         30 * time.Second,
	})
	if err != nil {
		t.Fatalf("new Clerk verifier: %v", err)
	}
	concrete, ok := verifier.(*clerkVerifier)
	if !ok {
		t.Fatalf("verifier type = %T, want *clerkVerifier", verifier)
	}
	if concrete.clockSkew != 30*time.Second {
		t.Fatalf("clock skew = %s, want 30s", concrete.clockSkew)
	}
}

func TestNewClerkVerifierRejectsUnsafeClockSkew(t *testing.T) {
	t.Parallel()

	for _, clockSkew := range []time.Duration{-time.Second, 30*time.Second + time.Nanosecond} {
		_, err := NewClerkVerifier(ClerkVerifierConfig{
			JWTKey:            testPublicKey(t),
			AuthorizedParties: []string{"http://localhost:4200"},
			ClockSkew:         clockSkew,
		})
		if !errors.Is(err, ErrInvalidClerkVerifierConfig) {
			t.Fatalf("clock skew %s error = %v, want ErrInvalidClerkVerifierConfig", clockSkew, err)
		}
	}
}
