package listingmedia

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SupabaseStorageConfig is server-owned configuration, never a browser DTO.
type SupabaseStorageConfig struct {
	Origin     string
	ServerKey  string
	Bucket     string
	HTTPClient *http.Client
}

type supabaseStorage struct {
	origin, serverKey, bucket string
	client                    *http.Client
}

var storageBucketPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,62}$`)

func NewSupabaseStorage(config SupabaseStorageConfig) (ManagedStorage, error) {
	origin := strings.TrimSuffix(strings.TrimSpace(config.Origin), "/")
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost"))) || strings.TrimSpace(config.ServerKey) == "" || strings.ContainsAny(config.ServerKey, "\r\n") || !storageBucketPattern.MatchString(config.Bucket) {
		return nil, ErrUnavailable
	}
	client := http.Client{Timeout: 15 * time.Second}
	if config.HTTPClient != nil {
		client = *config.HTTPClient
	}
	if client.Timeout <= 0 || client.Timeout > 15*time.Second {
		client.Timeout = 15 * time.Second
	}
	// Even a same-host redirect could forward server credentials to another path.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return supabaseStorage{origin: origin, serverKey: config.ServerKey, bucket: config.Bucket, client: &client}, nil
}

func (s supabaseStorage) CreateUploadReservation(ctx context.Context, id uuid.UUID, input UploadRequest) (StorageReservation, error) {
	if id == uuid.Nil || !validUploadRequest(input) {
		return StorageReservation{}, ErrInvalidUpload
	}
	object := "pending/" + id.String()
	path := "/object/upload/sign/" + s.bucket + "/" + object
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.origin+"/storage/v1"+path, strings.NewReader(`{}`))
	if err != nil {
		return StorageReservation{}, ErrUnavailable
	}
	req.Header.Set("Authorization", "Bearer "+s.serverKey)
	req.Header.Set("apikey", s.serverKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-upsert", "false")
	response, err := s.client.Do(req)
	if err != nil {
		return StorageReservation{}, ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return StorageReservation{}, ErrUnavailable
	}
	const limit = 16 * 1024
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || len(body) > limit {
		return StorageReservation{}, ErrUnavailable
	}
	var result struct {
		URL string `json:"url"`
	}
	if json.Unmarshal(body, &result) != nil || strings.Contains(result.URL, s.serverKey) {
		return StorageReservation{}, ErrUnavailable
	}
	signed, err := url.Parse(result.URL)
	if err != nil || signed.IsAbs() || signed.Host != "" || signed.User != nil || signed.Path != path || signed.RawPath != "" || signed.Fragment != "" {
		return StorageReservation{}, ErrUnavailable
	}
	query, err := url.ParseQuery(signed.RawQuery)
	tokens := query["token"]
	if err != nil || len(query) != 1 || len(tokens) != 1 || tokens[0] == "" || len(tokens[0]) > 8192 || strings.ContainsAny(tokens[0], "\r\n\x00") {
		return StorageReservation{}, ErrUnavailable
	}
	return StorageReservation{ObjectReference: object, Capability: UploadCapability{URL: s.origin + "/storage/v1" + signed.String(), Method: http.MethodPut, Headers: map[string]string{"Content-Type": strings.ToLower(input.ContentType)}}}, nil
}
