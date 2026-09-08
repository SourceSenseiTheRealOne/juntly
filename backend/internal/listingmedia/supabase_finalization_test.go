package listingmedia

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func finalizationAdapter(t *testing.T, handler http.HandlerFunc) FinalizationStorage {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	storage, err := NewSupabaseStorage(SupabaseStorageConfig{Origin: server.URL, ServerKey: "unit-test-key", Bucket: "listing-images"})
	if err != nil {
		t.Fatal(err)
	}
	adapter, ok := storage.(FinalizationStorage)
	if !ok {
		t.Fatal("Supabase adapter does not support finalization")
	}
	return adapter
}

func TestSupabaseFinalizationPrivateRoundTripAndImmutableReplay(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	body := testPNG(t, 2, 3)
	verified, err := SanitizeImage(body, imageRequest(body))
	if err != nil {
		t.Fatal(err)
	}
	pending := "pending/" + id.String()
	reference := "verified/" + id.String() + "/" + verified.ChecksumSHA256 + ".png"
	objects := map[string][]byte{pending: body}
	writes, reads := 0, 0
	storage := finalizationAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer unit-test-key" || r.Header.Get("apikey") != "unit-test-key" {
			t.Error("missing server authentication")
			w.WriteHeader(401)
			return
		}
		if r.Method == http.MethodGet {
			reads++
			key := strings.TrimPrefix(r.URL.Path, "/storage/v1/object/authenticated/listing-images/")
			data, ok := objects[key]
			if !ok {
				w.WriteHeader(404)
				return
			}
			_, _ = w.Write(data)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/storage/v1/object/listing-images/"+reference || r.Header.Get("x-upsert") != "false" || r.Header.Get("Content-Type") != "image/png" {
			t.Error("unsafe object write")
			w.WriteHeader(400)
			return
		}
		writes++
		if _, exists := objects[reference]; exists {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"statusCode":"409","error":"Duplicate","message":"The resource already exists"}`))
			return
		}
		data, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Error(readErr)
		}
		objects[reference] = data
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"Key":"listing-images/verified/image.png"}`))
	})
	got, err := storage.DownloadPending(ctx, pending)
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("private download failed: %v", err)
	}
	for range 2 {
		ref, err := storage.WriteVerified(ctx, id, verified)
		if err != nil || ref != reference {
			t.Fatalf("publication/replay failed: %v", err)
		}
	}
	if writes != 2 || reads != 3 || !bytes.Equal(objects[reference], verified.Bytes) {
		t.Fatal("immutable write did not verify stored bytes on both attempts")
	}
}

func TestSupabaseFinalizationRejectsUnsafeReferencesBeforeIO(t *testing.T) {
	calls := 0
	storage := finalizationAdapter(t, func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) })
	for _, ref := range []string{"", "pending/../key", "pending/%2e%2e", "https://evil.example", "pending/" + uuid.Nil.String(), "pending/" + uuid.New().String() + "?token=bad", "verified/" + uuid.New().String()} {
		if _, err := storage.DownloadPending(context.Background(), ref); !errors.Is(err, ErrInvalidUpload) {
			t.Fatalf("unsafe reference accepted: %q", ref)
		}
	}
	if _, err := storage.WriteVerified(context.Background(), uuid.Nil, SanitizedImage{}); !errors.Is(err, ErrInvalidUpload) {
		t.Fatal("invalid publication accepted")
	}
	if calls != 0 {
		t.Fatal("invalid reference reached provider")
	}
}

func TestSupabaseFinalizationRejectsOversizeRedirectAndCorruptReadback(t *testing.T) {
	body := testPNG(t, 2, 3)
	verified, err := SanitizeImage(body, imageRequest(body))
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"oversize", "redirect", "corrupt", "provider-error"} {
		t.Run(mode, func(t *testing.T) {
			storage := finalizationAdapter(t, func(w http.ResponseWriter, r *http.Request) {
				switch mode {
				case "oversize":
					_, _ = w.Write(bytes.Repeat([]byte{1}, 10485761))
				case "redirect":
					w.Header().Set("Location", "/leaked-key")
					w.WriteHeader(302)
				case "corrupt":
					if r.Method == http.MethodPost {
						w.WriteHeader(200)
					} else {
						_, _ = w.Write([]byte("wrong bytes"))
					}
				default:
					w.WriteHeader(403)
					_, _ = w.Write([]byte("unit-test-key"))
				}
			})
			if mode == "corrupt" {
				_, err = storage.WriteVerified(context.Background(), uuid.New(), verified)
			} else {
				_, err = storage.DownloadPending(context.Background(), "pending/"+uuid.New().String())
			}
			if !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), "unit-test-key") {
				t.Fatal("unsafe response accepted or provider error leaked")
			}
		})
	}
}
