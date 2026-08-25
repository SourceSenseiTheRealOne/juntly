package main

import (
	"encoding/base64"
	"testing"
)

func TestLoadRuntimeConfigBuildsOptionalServerOnlyContactCipher(t *testing.T) {
	t.Parallel()
	key := make([]byte, 32)
	config, err := loadRuntimeConfig(func(name string) string {
		switch name {
		case "DATABASE_URL":
			return "postgresql://synthetic"
		case "CLERK_SECRET_KEY":
			return "synthetic-secret"
		case "CLERK_AUTHORIZED_PARTIES":
			return "http://localhost:4200"
		case "JUNTLY_CONTACT_ENCRYPTION_KEY":
			return base64.StdEncoding.EncodeToString(key)
		default:
			return ""
		}
	})
	if err != nil || config.contactCipher == nil {
		t.Fatalf("config/cipher = %#v/%v", config, err)
	}
	_, err = loadRuntimeConfig(func(name string) string {
		switch name {
		case "DATABASE_URL":
			return "postgresql://synthetic"
		case "CLERK_SECRET_KEY":
			return "synthetic-secret"
		case "CLERK_AUTHORIZED_PARTIES":
			return "http://localhost:4200"
		case "JUNTLY_CONTACT_ENCRYPTION_KEY":
			return "invalid"
		default:
			return ""
		}
	})
	if err == nil {
		t.Fatal("invalid contact key accepted")
	}
}
