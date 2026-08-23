package authn

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	clerk "github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
)

var (
	ErrInvalidClerkVerifierConfig = errors.New("invalid Clerk verifier configuration")
	ErrUnverifiedClerkToken       = errors.New("unverified Clerk token")
)

type ClerkVerifierConfig struct {
	SecretKey         string
	JWTKey            string
	AuthorizedParties []string
	ClockSkew         time.Duration
}

type clerkVerifier struct {
	jwk               *clerk.JSONWebKey
	jwksClient        *jwks.Client
	authorizedParties map[string]struct{}
	clockSkew         time.Duration
}

func NewClerkVerifier(config ClerkVerifierConfig) (Verifier, error) {
	parties := authorizedPartySet(config.AuthorizedParties)
	if len(parties) == 0 {
		return nil, ErrInvalidClerkVerifierConfig
	}
	if config.ClockSkew < 0 || config.ClockSkew > 30*time.Second {
		return nil, ErrInvalidClerkVerifierConfig
	}

	verifier := &clerkVerifier{
		authorizedParties: parties,
		clockSkew:         config.ClockSkew,
	}
	if jwtKey := strings.TrimSpace(config.JWTKey); jwtKey != "" {
		params := &clerkhttp.AuthorizationParams{}
		if err := clerkhttp.JSONWebKey(jwtKey)(params); err != nil {
			return nil, ErrInvalidClerkVerifierConfig
		}
		verifier.jwk = params.JWK
		return verifier, nil
	}

	secretKey := strings.TrimSpace(config.SecretKey)
	if secretKey == "" {
		return nil, ErrInvalidClerkVerifierConfig
	}
	verifier.jwksClient = jwks.NewClient(&clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{Key: &secretKey},
	})
	return verifier, nil
}

func (v *clerkVerifier) Verify(ctx context.Context, token string) (users.VerifiedIdentity, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{
		Token:                  token,
		JWK:                    v.jwk,
		JWKSClient:             v.jwksClient,
		AuthorizedPartyHandler: v.authorizedParty,
		Leeway:                 v.clockSkew,
	})
	if err != nil || claims == nil || claims.SessionID == "" || strings.TrimSpace(claims.Subject) == "" {
		return users.VerifiedIdentity{}, ErrUnverifiedClerkToken
	}

	return users.VerifiedIdentity{Subject: claims.Subject}, nil
}

func (v *clerkVerifier) authorizedParty(value string) bool {
	_, ok := v.authorizedParties[value]
	return ok
}

func authorizedPartySet(values []string) map[string]struct{} {
	parties := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			parties[value] = struct{}{}
		}
	}
	return parties
}
