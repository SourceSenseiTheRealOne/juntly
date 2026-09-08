package listingmedia

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image/png"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("listing media not found")

type ReadScope string

const (
	PublicRead    ReadScope = "public"
	OwnerRead     ReadScope = "owner"
	ModeratorRead ReadScope = "moderator"
)

// VerifiedMedia stays private to repository/storage adapters.
type VerifiedMedia struct {
	ID              uuid.UUID
	Ordinal         int
	ObjectReference string
	ByteSize        int64
	ChecksumSHA256  string
	Width, Height   int
}

// Photo is the allowlisted response; it contains no storage capability or key.
type Photo struct {
	ID      uuid.UUID `json:"id"`
	Ordinal int       `json:"ordinal"`
	Width   int       `json:"width"`
	Height  int       `json:"height"`
}

type ReadRepository interface {
	ListVerified(context.Context, ReadScope, uuid.UUID, uuid.UUID) ([]VerifiedMedia, error)
}
type VerifiedStorage interface {
	DownloadVerified(context.Context, VerifiedMedia) ([]byte, error)
}

type Reader struct {
	identities users.Service
	repository ReadRepository
	storage    VerifiedStorage
	slots      chan struct{}
}

func NewReader(identities users.Service, repository ReadRepository, storage VerifiedStorage) *Reader {
	return &Reader{identities: identities, repository: repository, storage: storage, slots: make(chan struct{}, 8)}
}
func (r *Reader) entries(ctx context.Context, identity users.VerifiedIdentity, scope ReadScope, listingID uuid.UUID) ([]VerifiedMedia, error) {
	if r == nil || r.repository == nil || r.storage == nil {
		return nil, ErrUnavailable
	}
	if listingID == uuid.Nil || (scope != PublicRead && scope != OwnerRead && scope != ModeratorRead) {
		return nil, ErrInvalidUpload
	}
	actor := uuid.Nil
	if scope != PublicRead {
		if r.identities == nil {
			return nil, ErrUnavailable
		}
		user, _, err := r.identities.Reconcile(ctx, identity)
		if errors.Is(err, users.ErrInvalidIdentity) {
			return nil, provideraccess.ErrUnauthorized
		}
		if err != nil || user.ID == uuid.Nil {
			return nil, ErrUnavailable
		}
		actor = user.ID
	}
	entries, err := r.repository.ListVerified(ctx, scope, actor, listingID)
	if err != nil {
		return nil, err
	}
	if len(entries) > 10 {
		return nil, ErrUnavailable
	}
	for _, entry := range entries {
		if !validVerifiedMedia(entry) {
			return nil, ErrUnavailable
		}
	}
	return entries, nil
}
func (r *Reader) List(ctx context.Context, identity users.VerifiedIdentity, scope ReadScope, listingID uuid.UUID) ([]Photo, error) {
	entries, err := r.entries(ctx, identity, scope, listingID)
	if err != nil {
		return nil, err
	}
	photos := make([]Photo, len(entries))
	for i, entry := range entries {
		photos[i] = Photo{ID: entry.ID, Ordinal: entry.Ordinal, Width: entry.Width, Height: entry.Height}
	}
	return photos, nil
}
func (r *Reader) Get(ctx context.Context, identity users.VerifiedIdentity, scope ReadScope, listingID, mediaID uuid.UUID) ([]byte, error) {
	if mediaID == uuid.Nil {
		return nil, ErrInvalidUpload
	}
	entries, err := r.entries(ctx, identity, scope, listingID)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.ID != mediaID {
			continue
		}
		select {
		case r.slots <- struct{}{}:
			defer func() { <-r.slots }()
		default:
			return nil, ErrUnavailable
		}
		body, err := r.storage.DownloadVerified(ctx, entry)
		if err != nil || !matchesVerifiedBytes(body, entry) {
			return nil, ErrUnavailable
		}
		return body, nil
	}
	return nil, ErrNotFound
}
func validVerifiedMedia(entry VerifiedMedia) bool {
	return entry.ID != uuid.Nil && validUploadRequest(UploadRequest{Ordinal: entry.Ordinal, ContentType: "image/png", ByteSize: entry.ByteSize, ChecksumSHA256: entry.ChecksumSHA256}) && entry.ObjectReference == verifiedReference(entry.ID, entry.ChecksumSHA256) && entry.Width > 0 && entry.Height > 0 && entry.Width <= 8192 && entry.Height <= 8192 && int64(entry.Width)*int64(entry.Height) <= 16_000_000
}
func matchesVerifiedBytes(body []byte, entry VerifiedMedia) bool {
	if !validVerifiedMedia(entry) || int64(len(body)) != entry.ByteSize {
		return false
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != entry.ChecksumSHA256 {
		return false
	}
	config, err := png.DecodeConfig(bytes.NewReader(body))
	return err == nil && config.Width == entry.Width && config.Height == entry.Height
}
