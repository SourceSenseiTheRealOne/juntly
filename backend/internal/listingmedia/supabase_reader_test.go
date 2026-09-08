package listingmedia

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestSupabaseVerifiedReaderChecksReferenceAndStoredImage(t *testing.T) {
	body := testPNG(t, 2, 3)
	image, err := SanitizeImage(body, imageRequest(body))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	entry := VerifiedMedia{ID: id, Ordinal: 1, ObjectReference: verifiedReference(id, image.ChecksumSHA256), ByteSize: int64(len(image.Bytes)), ChecksumSHA256: image.ChecksumSHA256, Width: image.Width, Height: image.Height}
	calls := 0
	response := image.Bytes
	storage := finalizationAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/storage/v1/object/authenticated/listing-images/"+entry.ObjectReference || r.Header.Get("Authorization") != "Bearer unit-test-key" {
			t.Error("wrong authenticated verified read")
			w.WriteHeader(403)
			return
		}
		_, _ = w.Write(response)
	})
	reader, ok := storage.(VerifiedStorage)
	if !ok {
		t.Fatal("storage does not support verified reads")
	}
	got, err := reader.DownloadVerified(context.Background(), entry)
	if err != nil || !bytes.Equal(got, image.Bytes) {
		t.Fatal("verified read failed")
	}
	response = append(image.Bytes, 1)
	if _, err := reader.DownloadVerified(context.Background(), entry); !errors.Is(err, ErrUnavailable) {
		t.Fatal("corrupt object accepted")
	}
	before := calls
	entry.ObjectReference = "pending/" + id.String()
	if _, err := reader.DownloadVerified(context.Background(), entry); !errors.Is(err, ErrInvalidUpload) || calls != before {
		t.Fatal("pending reference reached provider")
	}
}
