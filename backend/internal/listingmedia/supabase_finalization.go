package listingmedia

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const maxImageBytes = 10 * 1024 * 1024

var _ FinalizationStorage = supabaseStorage{}

func (s supabaseStorage) DownloadPending(ctx context.Context, reference string) ([]byte, error) {
	if !strings.HasPrefix(reference, "pending/") || !canonicalMediaID(strings.TrimPrefix(reference, "pending/")) {
		return nil, ErrInvalidUpload
	}
	return s.downloadObject(ctx, reference)
}

func canonicalMediaID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func (s supabaseStorage) downloadObject(ctx context.Context, reference string) ([]byte, error) {
	response, err := s.objectRequest(ctx, http.MethodGet, "/object/authenticated/"+s.bucket+"/"+reference, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.ContentLength > maxImageBytes {
		return nil, ErrUnavailable
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxImageBytes+1))
	if err != nil || len(body) == 0 || len(body) > maxImageBytes {
		return nil, ErrUnavailable
	}
	return body, nil
}

// WriteVerified never overwrites. Both new writes and duplicate retries must
// read back exactly the sanitized bytes before persistence can mark media ready.
func (s supabaseStorage) WriteVerified(ctx context.Context, id uuid.UUID, image SanitizedImage) (string, error) {
	if id == uuid.Nil || !validSanitizedImage(image) {
		return "", ErrInvalidUpload
	}
	reference := "verified/" + id.String() + "/" + image.ChecksumSHA256 + ".png"
	response, err := s.objectRequest(ctx, http.MethodPost, "/object/"+s.bucket+"/"+reference, image.Bytes)
	if err != nil {
		return "", err
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 16*1024+1))
	_ = response.Body.Close()
	if readErr != nil || len(body) > 16*1024 {
		return "", ErrUnavailable
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		var failure struct {
			StatusCode string `json:"statusCode"`
			Error      string `json:"error"`
		}
		duplicate := json.Unmarshal(body, &failure) == nil && failure.StatusCode == "409" && failure.Error == "Duplicate"
		if response.StatusCode != http.StatusConflict && !(response.StatusCode == http.StatusBadRequest && duplicate) {
			return "", ErrUnavailable
		}
	}
	stored, err := s.downloadObject(ctx, reference)
	if err != nil || !bytes.Equal(stored, image.Bytes) {
		return "", ErrUnavailable
	}
	return reference, nil
}

func validSanitizedImage(image SanitizedImage) bool {
	if image.ContentType != "image/png" || len(image.Bytes) == 0 || len(image.Bytes) > maxImageBytes || image.Width < 1 || image.Height < 1 || image.Width > 8192 || image.Height > 8192 || int64(image.Width)*int64(image.Height) > 16_000_000 {
		return false
	}
	sum := sha256.Sum256(image.Bytes)
	return image.ChecksumSHA256 == hex.EncodeToString(sum[:])
}

func (s supabaseStorage) objectRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.origin+"/storage/v1"+path, bytes.NewReader(body))
	if err != nil {
		return nil, ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+s.serverKey)
	req.Header.Set("apikey", s.serverKey)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "image/png")
		req.Header.Set("x-upsert", "false")
		req.Header.Set("Cache-Control", "no-store")
	}
	response, err := s.client.Do(req)
	if err != nil {
		return nil, ErrUnavailable
	}
	return response, nil
}
