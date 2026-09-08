package listingmedia

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidUpload = errors.New("invalid listing media upload")
	ErrUnavailable   = errors.New("listing media unavailable")
	ErrConflict      = errors.New("listing media revision conflict")
)

type UploadRequest struct {
	Ordinal        int
	ContentType    string
	ByteSize       int64
	ChecksumSHA256 string
}
type UploadCapability struct {
	URL     string
	Method  string
	Headers map[string]string
}
type StorageReservation struct {
	ObjectReference string
	Capability      UploadCapability
}
type UploadIntent struct {
	MediaID    uuid.UUID
	Capability UploadCapability
}
