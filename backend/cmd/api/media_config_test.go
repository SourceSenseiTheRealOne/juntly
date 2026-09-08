package main

import "testing"

func TestMediaRuntimeConfigurationIsAllOrNothing(t *testing.T) {
	base := map[string]string{"DATABASE_URL": "postgresql://synthetic", "CLERK_SECRET_KEY": "synthetic-secret", "CLERK_AUTHORIZED_PARTIES": "http://localhost:4200"}
	storage := map[string]string{"JUNTLY_STORAGE_ORIGIN": "https://storage.example.test", "JUNTLY_STORAGE_SERVER_KEY": "unit-test-key", "JUNTLY_STORAGE_BUCKET": "listing-images"}
	for _, missing := range []string{"all", "", "JUNTLY_STORAGE_ORIGIN", "JUNTLY_STORAGE_SERVER_KEY", "JUNTLY_STORAGE_BUCKET"} {
		t.Run(missing, func(t *testing.T) {
			config, err := loadRuntimeConfig(func(key string) string {
				if v, ok := base[key]; ok {
					return v
				}
				if missing == "all" || key == missing {
					return ""
				}
				return storage[key]
			})
			if missing != "" && missing != "all" {
				if err == nil {
					t.Fatal("partial storage config accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if (config.mediaStorage != nil) != (missing == "") {
				t.Fatal("storage activation did not match complete explicit configuration")
			}
		})
	}
}
