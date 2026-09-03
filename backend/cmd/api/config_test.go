package main

import (
	"testing"
)

func TestLoadRuntimeConfigRejectsMissingDatabaseURL(t *testing.T) {
	t.Parallel()

	_, err := loadRuntimeConfig(func(key string) string {
		if key == "CLERK_SECRET_KEY" {
			return "synthetic-secret"
		}
		if key == "CLERK_AUTHORIZED_PARTIES" {
			return "http://localhost:4200"
		}
		return ""
	})
	if err == nil {
		t.Fatal("error = nil, want missing database URL rejection")
	}
}

func TestLoadRuntimeConfigRejectsMissingAuthorizedParties(t *testing.T) {
	t.Parallel()

	_, err := loadRuntimeConfig(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgresql://synthetic"
		case "CLERK_SECRET_KEY":
			return "synthetic-secret"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("error = nil, want missing authorized-party rejection")
	}
}

func TestLoadRuntimeConfigBuildsVerifierWithoutPersistingCredentials(t *testing.T) {
	t.Parallel()

	config, err := loadRuntimeConfig(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgresql://synthetic"
		case "CLERK_SECRET_KEY":
			return "synthetic-secret"
		case "CLERK_AUTHORIZED_PARTIES":
			return "http://localhost:4200, https://app.example.test"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("load runtime config: %v", err)
	}
	if config.databaseURL != "postgresql://synthetic" {
		t.Fatalf("database URL = %q", config.databaseURL)
	}
	if config.verifier == nil {
		t.Fatal("verifier is nil")
	}
}

func TestLoadRuntimeConfigAcceptsBoundedClerkClockSkew(t *testing.T) {
	t.Parallel()

	_, err := loadRuntimeConfig(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return "postgresql://synthetic"
		case "CLERK_SECRET_KEY":
			return "synthetic-secret"
		case "CLERK_AUTHORIZED_PARTIES":
			return "http://localhost:4200"
		case "CLERK_CLOCK_SKEW":
			return "30s"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("load runtime config: %v", err)
	}
}

func TestLoadRuntimeConfigRejectsInvalidClerkClockSkew(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"not-a-duration", "-1s", "30.000000001s"} {
		value := value
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			_, err := loadRuntimeConfig(func(key string) string {
				switch key {
				case "DATABASE_URL":
					return "postgresql://synthetic"
				case "CLERK_SECRET_KEY":
					return "synthetic-secret"
				case "CLERK_AUTHORIZED_PARTIES":
					return "http://localhost:4200"
				case "CLERK_CLOCK_SKEW":
					return value
				default:
					return ""
				}
			})
			if err == nil {
				t.Fatal("error = nil, want invalid clock-skew rejection")
			}
		})
	}
}

func TestLoadRuntimeConfigEnablesStripeOnlyWithCompleteServerConfiguration(t *testing.T) {
	t.Parallel()
	config, err := loadRuntimeConfig(func(key string) string {
		return map[string]string{
			"DATABASE_URL":             "postgresql://synthetic",
			"CLERK_SECRET_KEY":         "synthetic-secret",
			"CLERK_AUTHORIZED_PARTIES": "http://localhost:4200",
			"STRIPE_SECRET_KEY":        "sk_test_synthetic",
			"STRIPE_WEBHOOK_SECRET":    "whsec_synthetic",
			"JUNTLY_PUBLIC_ORIGIN":     "https://vila.example",
			"JUNTLY_PLATFORM_FEE_BPS":  "1000",
		}[key]
	})
	if err != nil {
		t.Fatalf("load runtime config: %v", err)
	}
	if config.paymentGateway == nil || config.platformFeeBPS != 1000 {
		t.Fatalf("payment config = %#v/%d", config.paymentGateway, config.platformFeeBPS)
	}
}

func TestLoadRuntimeConfigRejectsPartialStripeConfiguration(t *testing.T) {
	t.Parallel()
	_, err := loadRuntimeConfig(func(key string) string {
		return map[string]string{
			"DATABASE_URL":             "postgresql://synthetic",
			"CLERK_SECRET_KEY":         "synthetic-secret",
			"CLERK_AUTHORIZED_PARTIES": "http://localhost:4200",
			"STRIPE_SECRET_KEY":        "sk_test_synthetic",
		}[key]
	})
	if err == nil {
		t.Fatal("partial Stripe configuration accepted")
	}
}
