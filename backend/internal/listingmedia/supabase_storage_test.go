package listingmedia

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestSupabaseStorageMintsOnlyScopedUploadCapability(t *testing.T) {
	id := uuid.New()
	path := "/object/upload/sign/vila-quarantine/pending/" + id.String()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/storage/v1"+path {
			t.Errorf("unexpected storage route %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer synthetic-server-key" || r.Header.Get("apikey") != "synthetic-server-key" || r.Header.Get("x-upsert") != "false" {
			t.Error("missing server authentication or overwrite restriction")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"url": path + "?token=synthetic-upload-capability"})
	}))
	defer server.Close()
	storage, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: server.URL, ServerKey: "synthetic-server-key", Bucket: "vila-quarantine"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := storage.CreateUploadReservation(context.Background(), id, validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if r.ObjectReference != "pending/"+id.String() || r.Capability.URL != server.URL+"/storage/v1"+path+"?token=synthetic-upload-capability" || r.Capability.Method != "PUT" || len(r.Capability.Headers) != 1 || r.Capability.Headers["Content-Type"] != "image/webp" {
		t.Fatal("unexpected capability projection")
	}
	if strings.Contains(r.Capability.URL, "synthetic-server-key") {
		t.Fatal("server key exposed")
	}
	if _, err := storage.CreateUploadReservation(context.Background(), uuid.Nil, validRequest()); !errors.Is(err, ErrInvalidUpload) {
		t.Fatal("nil media ID accepted")
	}
	if calls != 1 {
		t.Fatalf("invalid upload performed network request: %d calls", calls)
	}
}

func TestSupabaseStorageRejectsUnsafeConfig(t *testing.T) {
	for _, origin := range []string{"http://storage.example.com", "https://user:pass@storage.example.com", "https://storage.example.com/path", "https://storage.example.com?x=1", "https://storage.example.com#fragment", "file:///tmp"} {
		if _, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: origin, ServerKey: "synthetic-server-key", Bucket: "vila-quarantine"}); !errors.Is(err, ErrUnavailable) {
			t.Errorf("accepted unsafe origin %q", origin)
		}
	}
	for _, bucket := range []string{"", "../other", "bucket/path", "bucket?query", "Bucket"} {
		if _, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: "https://storage.example.com", ServerKey: "synthetic-server-key", Bucket: bucket}); err == nil {
			t.Errorf("accepted unsafe bucket %q", bucket)
		}
	}
	if _, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: "https://storage.example.com", Bucket: "vila-quarantine"}); err == nil {
		t.Fatal("accepted missing server key")
	}
}

func TestSupabaseStorageRejectsUntrustedCapabilityResponses(t *testing.T) {
	for _, response := range []string{
		`{"url":"https://other.example.com/upload?token=abc"}`,
		`{"url":"//other.example.com/upload?token=abc"}`,
		`{"url":"/object/upload/sign/other/path?token=abc"}`,
		`{"url":"/object/upload/sign/vila-quarantine/pending/ID"}`,
		`{"url":"/object/upload/sign/vila-quarantine/pending/ID?token=a&token=b"}`,
		`{"url":"/object/upload/sign/vila-quarantine/pending/ID?token=synthetic-server-key"}`,
		`{"url":"/object/upload/sign/vila-quarantine/pending/ID?token=abc#fragment"}`,
		strings.Repeat("x", 17000),
	} {
		id := uuid.New()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(strings.ReplaceAll(response, "ID", id.String())))
		}))
		storage, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: server.URL, ServerKey: "synthetic-server-key", Bucket: "vila-quarantine"})
		if err != nil {
			t.Fatal(err)
		}
		r, err := storage.CreateUploadReservation(context.Background(), id, validRequest())
		server.Close()
		if !errors.Is(err, ErrUnavailable) || r.Capability.URL != "" {
			t.Fatal("unsafe provider response became capability")
		}
	}
}

func TestSupabaseStorageDoesNotFollowCredentialedRedirect(t *testing.T) {
	followed := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { followed = true }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	storage, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: server.URL, ServerKey: "synthetic-server-key", Bucket: "vila-quarantine", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = storage.CreateUploadReservation(context.Background(), uuid.New(), validRequest())
	if !errors.Is(err, ErrUnavailable) || followed {
		t.Fatal("credentialed redirect followed")
	}
}
