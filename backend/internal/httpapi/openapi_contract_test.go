package httpapi_test

import (
	"os"
	"strings"
	"testing"
)

func TestReconcileOpenAPIContract(t *testing.T) {
	t.Parallel()

	contract, err := os.ReadFile("../../../openapi/juntly-api.v1.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contents := string(contract)

	for _, required := range []string{
		"/api/v1/auth/reconcile:",
		"post:",
		"operationId: reconcileInternalUser",
		"clerkSession: []",
		"InternalUserResponse:",
		"format: uuid",
		"createdAt:",
		"format: date-time",
		"UNAUTHORIZED",
		"SERVICE_UNAVAILABLE",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("OpenAPI contract does not contain %q", required)
		}
	}

	operation := contents[strings.Index(contents, "/api/v1/auth/reconcile:"):]
	if strings.Contains(operation, "requestBody:") {
		t.Fatal("reconciliation operation must not accept a request body")
	}
}

func TestAccountCapabilitiesOpenAPIContract(t *testing.T) {
	t.Parallel()

	contract, err := os.ReadFile("../../../openapi/juntly-api.v1.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contents := string(contract)

	for _, required := range []string{
		"/api/v1/me/account:",
		"operationId: getAccountCapabilities",
		"operationId: updateAccountCapabilities",
		"AccountCapabilitiesResponse:",
		"UpdateAccountCapabilitiesRequest:",
		"customerEnabled:",
		"const: true",
		"providerEnabled:",
		"onboardingCompletedAt:",
		"INVALID_REQUEST",
		"clerkSession: []",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("OpenAPI contract does not contain %q", required)
		}
	}

	accountPath := contents[strings.Index(contents, "/api/v1/me/account:"):]
	if !strings.Contains(accountPath, "get:") || !strings.Contains(accountPath, "put:") {
		t.Fatal("account contract must declare GET and PUT")
	}
}
