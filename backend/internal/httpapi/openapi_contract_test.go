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

	operation := openAPIPathBlock(t, contents, "/api/v1/auth/reconcile:")
	if strings.Contains(operation, "requestBody:") {
		t.Fatal("reconciliation operation must not accept a request body")
	}
}

func openAPIPathBlock(t *testing.T, contents, path string) string {
	t.Helper()
	start := strings.Index(contents, path)
	if start < 0 {
		t.Fatalf("OpenAPI contract does not contain path %q", path)
	}
	remainder := contents[start:]
	end := len(remainder)
	for _, boundary := range []string{"\n  /api/", "\ncomponents:"} {
		if index := strings.Index(remainder[1:], boundary); index >= 0 && index+1 < end {
			end = index + 1
		}
	}
	return remainder[:end]
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

func TestTaxonomyLocationsProviderProfileOpenAPIContract(t *testing.T) {
	t.Parallel()

	contract, err := os.ReadFile("../../../openapi/juntly-api.v1.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contents := string(contract)
	for _, required := range []string{
		"/api/v1/catalog/categories:",
		"operationId: listServiceCategories",
		"/api/v1/reference/languages:",
		"operationId: listSpokenLanguages",
		"/api/v1/reference/localities:",
		"operationId: listLocalities",
		"nearLocalityId",
		"radiusKm",
		"/api/v1/me/provider-profile:",
		"operationId: getProviderProfile",
		"operationId: replaceProviderProfile",
		"ProviderProfileResponse:",
		"ReplaceProviderProfileRequest:",
		"serviceLocalityIds:",
		"languageCodes:",
		"FORBIDDEN",
		"OpenStreetMap contributors",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("OpenAPI contract does not contain %q", required)
		}
	}
	for _, prohibited := range []string{"phoneNumber:", "whatsapp:", "exactAddress:", "latitude:", "longitude:", "internalUserId:"} {
		if strings.Contains(contents, prohibited) {
			t.Fatalf("OpenAPI contract must not contain %q", prohibited)
		}
	}
}

func TestListingsModerationMediaOpenAPIContract(t *testing.T) {
	t.Parallel()
	contract, err := os.ReadFile("../../../openapi/juntly-api.v1.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	contents := string(contract)
	for _, required := range []string{
		"/api/v1/me/listings:", "operationId: listMyListings", "operationId: createListing",
		"/api/v1/me/listings/{listingId}:", "operationId: getMyListing", "operationId: replaceMyDraftListing",
		"/api/v1/me/listings/{listingId}/submit:", "operationId: submitListingForReview",
		"/api/v1/me/listings/{listingId}/pause:", "operationId: pauseListing",
		"/api/v1/me/listings/{listingId}/archive:", "operationId: archiveListing",
		"/api/v1/me/listings/{listingId}/media/upload-intents:", "operationId: createListingMediaUploadIntent",
		"/api/v1/moderation/listings:", "operationId: listPendingModerationListings", "operationId: approveListing", "operationId: rejectListing",
		"ListingResponse:", "CreateListingRequest:", "UploadIntentResponse:", "CONFLICT", "clerkSession: []",
	} {
		if !strings.Contains(contents, required) {
			t.Fatalf("OpenAPI contract does not contain %q", required)
		}
	}
	for _, prohibited := range []string{"phoneNumber:", "whatsapp:", "exactAddress:", "objectReference:", "accessKey:", "bucket:"} {
		if strings.Contains(contents, prohibited) {
			t.Fatalf("OpenAPI contract must not contain %q", prohibited)
		}
	}
}
