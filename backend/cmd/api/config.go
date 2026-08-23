package main

import (
	"errors"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
)

var ErrInvalidRuntimeConfig = errors.New("invalid API runtime configuration")

type runtimeConfig struct {
	databaseURL string
	verifier    authn.Verifier
}

func loadRuntimeConfig(lookup func(string) string) (runtimeConfig, error) {
	databaseURL := strings.TrimSpace(lookup("DATABASE_URL"))
	if databaseURL == "" {
		return runtimeConfig{}, ErrInvalidRuntimeConfig
	}
	clockSkew, err := parseOptionalDuration(lookup("CLERK_CLOCK_SKEW"))
	if err != nil {
		return runtimeConfig{}, ErrInvalidRuntimeConfig
	}

	verifier, err := authn.NewClerkVerifier(authn.ClerkVerifierConfig{
		SecretKey:         lookup("CLERK_SECRET_KEY"),
		JWTKey:            lookup("CLERK_JWT_KEY"),
		AuthorizedParties: strings.Split(lookup("CLERK_AUTHORIZED_PARTIES"), ","),
		ClockSkew:         clockSkew,
	})
	if err != nil {
		return runtimeConfig{}, ErrInvalidRuntimeConfig
	}

	return runtimeConfig{databaseURL: databaseURL, verifier: verifier}, nil
}

func parseOptionalDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return time.ParseDuration(value)
}
