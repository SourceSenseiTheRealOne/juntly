package discovery

import (
	"reflect"
	"strings"
	"testing"
)

func TestPublicListingProjectionExcludesPrivateColumnsAndFields(t *testing.T) {
	t.Parallel()
	columns := strings.ToLower(publicListingColumns)
	for _, forbidden := range []string{
		"internal_user_id",
		"clerk_subject",
		"object_reference",
		"checksum_sha256",
		"listing_event",
		"reason",
		"latitude",
		"longitude",
		"center",
		"bio",
		"service_localit",
		"phone",
		"email",
	} {
		if strings.Contains(columns, forbidden) {
			t.Fatalf("public select contains private column %q", forbidden)
		}
	}
	allowed := map[string]bool{
		"ID": true, "Title": true, "Description": true,
		"CategoryID": true, "CategorySlug": true, "CategoryName": true,
		"PrimaryLocalityID": true, "LocalitySlug": true, "LocalityName": true,
		"PriceType": true, "PriceMinor": true, "Currency": true,
		"TravelsToCustomer": true, "ReceivesCustomer": true, "RemoteServices": true,
		"ProviderDisplayName": true, "ProviderType": true, "UpdatedAt": true,
	}
	typeOfListing := reflect.TypeOf(Listing{})
	if typeOfListing.NumField() != len(allowed) {
		t.Fatalf("public field count = %d, want %d", typeOfListing.NumField(), len(allowed))
	}
	for index := 0; index < typeOfListing.NumField(); index++ {
		if !allowed[typeOfListing.Field(index).Name] {
			t.Fatalf("private public model field %q", typeOfListing.Field(index).Name)
		}
	}
}
