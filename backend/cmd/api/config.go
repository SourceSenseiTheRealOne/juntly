package main

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/authn"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/contactreveal"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/payments"
)

var ErrInvalidRuntimeConfig = errors.New("invalid API runtime configuration")

type runtimeConfig struct {
	databaseURL    string
	verifier       authn.Verifier
	contactCipher  contactreveal.Cipher
	paymentGateway payments.Gateway
	platformFeeBPS int
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
	var contactCipher contactreveal.Cipher
	if encodedKey := strings.TrimSpace(lookup("JUNTLY_CONTACT_ENCRYPTION_KEY")); encodedKey != "" {
		contactCipher, err = contactreveal.NewCipher(encodedKey)
		if err != nil {
			return runtimeConfig{}, ErrInvalidRuntimeConfig
		}
	}
	var paymentGateway payments.Gateway
	platformFeeBPS := 0
	stripeSecret := strings.TrimSpace(lookup("STRIPE_SECRET_KEY"))
	stripeWebhook := strings.TrimSpace(lookup("STRIPE_WEBHOOK_SECRET"))
	publicOrigin := strings.TrimSpace(lookup("JUNTLY_PUBLIC_ORIGIN"))
	feeValue := strings.TrimSpace(lookup("JUNTLY_PLATFORM_FEE_BPS"))
	if stripeSecret != "" || stripeWebhook != "" || publicOrigin != "" || feeValue != "" {
		if stripeSecret == "" || stripeWebhook == "" || publicOrigin == "" || feeValue == "" {
			return runtimeConfig{}, ErrInvalidRuntimeConfig
		}
		platformFeeBPS, err = strconv.Atoi(feeValue)
		if err != nil || platformFeeBPS < 0 || platformFeeBPS >= 10_000 {
			return runtimeConfig{}, ErrInvalidRuntimeConfig
		}
		apiBase := strings.TrimSpace(lookup("STRIPE_API_BASE"))
		if apiBase == "" {
			apiBase = "https://api.stripe.com"
		}
		paymentGateway, err = payments.NewStripeGateway(payments.StripeConfig{SecretKey: stripeSecret, WebhookSecret: stripeWebhook, APIBase: apiBase, PublicOrigin: publicOrigin, Now: time.Now})
		if err != nil {
			return runtimeConfig{}, ErrInvalidRuntimeConfig
		}
	}

	return runtimeConfig{databaseURL: databaseURL, verifier: verifier, contactCipher: contactCipher, paymentGateway: paymentGateway, platformFeeBPS: platformFeeBPS}, nil
}

func parseOptionalDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return time.ParseDuration(value)
}
