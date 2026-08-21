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
