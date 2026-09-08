package listingmedia

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/jpeg"
	"image/png"
	"strings"

	_ "golang.org/x/image/webp"
)

// SanitizedImage contains newly encoded pixels, never the original upload.
type SanitizedImage struct {
	Bytes          []byte
	ContentType    string
	ChecksumSHA256 string
	Width, Height  int
}

func SanitizeImage(body []byte, request UploadRequest) (SanitizedImage, error) {
	if !validUploadRequest(request) || int64(len(body)) != request.ByteSize {
		return SanitizedImage{}, ErrInvalidUpload
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != request.ChecksumSHA256 {
		return SanitizedImage{}, ErrInvalidUpload
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 8192 || config.Height > 8192 || int64(config.Width)*int64(config.Height) > 16_000_000 {
		return SanitizedImage{}, ErrInvalidUpload
	}
	mime := map[string]string{"jpeg": "image/jpeg", "png": "image/png", "webp": "image/webp"}[format]
	if mime == "" || !strings.EqualFold(mime, request.ContentType) {
		return SanitizedImage{}, ErrInvalidUpload
	}
	decoded, decodedFormat, err := image.Decode(bytes.NewReader(body))
	if err != nil || decodedFormat != format || decoded.Bounds().Dx() != config.Width || decoded.Bounds().Dy() != config.Height {
		return SanitizedImage{}, ErrInvalidUpload
	}
	output := boundedImageBuffer{}
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(&output, decoded); err != nil {
		return SanitizedImage{}, ErrInvalidUpload
	}
	sanitized := output.Bytes()
	sum = sha256.Sum256(sanitized)
	return SanitizedImage{Bytes: sanitized, ContentType: "image/png", ChecksumSHA256: hex.EncodeToString(sum[:]), Width: config.Width, Height: config.Height}, nil
}

type boundedImageBuffer struct{ bytes.Buffer }

func (b *boundedImageBuffer) Write(p []byte) (int, error) {
	if len(p) > 10485760-b.Len() {
		return 0, ErrInvalidUpload
	}
	return b.Buffer.Write(p)
}
