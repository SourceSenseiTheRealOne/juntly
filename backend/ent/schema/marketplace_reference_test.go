package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestMarketplaceReferenceSchemas(t *testing.T) {
	t.Parallel()

	contracts := map[string][]string{
		"supportedlocale.go": {
			"type SupportedLocale struct",
			"field.String(\"id\")",
			"MaxLen(10)",
			"field.Bool(\"active\")",
			"field.Int(\"sort_order\")",
			"Table: \"supported_locales\"",
		},
		"servicecategory.go": {
			"type ServiceCategory struct",
			"field.UUID(\"id\"",
			"Default(uuid.New)",
			"field.UUID(\"parent_id\"",
			"Optional()",
			"Nillable()",
			"field.String(\"slug\")",
			"Unique()",
			"field.Bool(\"active\")",
			"field.Int(\"sort_order\")",
			"Table: \"service_categories\"",
		},
		"servicecategorytranslation.go": {
			"type ServiceCategoryTranslation struct",
			"field.UUID(\"category_id\"",
			"field.String(\"locale\")",
			"field.String(\"name\")",
			"field.String(\"description\")",
			"field.ID(\"category_id\", \"locale\")",
			"Table: \"service_category_translations\"",
		},
		"spokenlanguage.go": {
			"type SpokenLanguage struct",
			"field.String(\"id\")",
			"field.Bool(\"active\")",
			"field.Int(\"sort_order\")",
			"Table: \"spoken_languages\"",
		},
		"spokenlanguagetranslation.go": {
			"type SpokenLanguageTranslation struct",
			"field.String(\"language_code\")",
			"field.String(\"locale\")",
			"field.String(\"name\")",
			"field.ID(\"language_code\", \"locale\")",
			"Table: \"spoken_language_translations\"",
		},
		"administrativearea.go": {
			"type AdministrativeArea struct",
			"field.UUID(\"id\"",
			"field.String(\"source\")",
			"field.String(\"source_version\")",
			"field.String(\"external_code\")",
			"field.String(\"kind\")",
			"field.String(\"name\")",
			"field.UUID(\"parent_id\"",
			"field.Bool(\"active\")",
			"Table: \"administrative_areas\"",
		},
		"locality.go": {
			"type Locality struct",
			"field.UUID(\"id\"",
			"field.String(\"slug\")",
			"field.String(\"name\")",
			"field.UUID(\"parent_parish_id\"",
			"field.String(\"source\")",
			"field.String(\"source_element_id\")",
			"field.Float(\"latitude\")",
			"field.Float(\"longitude\")",
			"field.Bool(\"active\")",
			"Table: \"localities\"",
		},
		"providerprofile.go": {
			"type ProviderProfile struct",
			"field.UUID(\"id\"",
			"StorageKey(\"internal_user_id\")",
			"field.String(\"display_name\")",
			"field.String(\"provider_type\")",
			"field.String(\"bio\")",
			"field.UUID(\"primary_locality_id\"",
			"field.Int(\"max_travel_distance_km\")",
			"field.Bool(\"travels_to_customer\")",
			"field.Bool(\"receives_customer\")",
			"field.Bool(\"remote_services\")",
			"Table: \"provider_profiles\"",
		},
		"providerservicelocality.go": {
			"type ProviderServiceLocality struct",
			"field.UUID(\"internal_user_id\"",
			"field.UUID(\"locality_id\"",
			"field.ID(\"internal_user_id\", \"locality_id\")",
			"Table: \"provider_service_localities\"",
		},
		"providerspokenlanguage.go": {
			"type ProviderSpokenLanguage struct",
			"field.UUID(\"internal_user_id\"",
			"field.String(\"language_code\")",
			"field.ID(\"internal_user_id\", \"language_code\")",
			"Table: \"provider_spoken_languages\"",
		},
	}

	for path, required := range contracts {
		path, required := path, required
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read schema %s: %v", path, err)
			}
			schema := string(contents)
			for _, requirement := range required {
				if !strings.Contains(schema, requirement) {
					t.Errorf("schema %s does not contain %q", path, requirement)
				}
			}
		})
	}
}

func TestProviderProfileSchemaExcludesPrivateAndDeferredFields(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("providerprofile.go")
	if err != nil {
		t.Fatalf("read provider profile schema: %v", err)
	}
	schema := strings.ToLower(string(contents))
	for _, prohibited := range []string{
		"email",
		"phone",
		"whatsapp",
		"address",
		"clerk",
		"token",
		"session",
		"verification",
		"payment",
		"listing",
		"review",
		"portfolio",
	} {
		if strings.Contains(schema, prohibited) {
			t.Errorf("provider profile schema must not include %q", prohibited)
		}
	}
}
