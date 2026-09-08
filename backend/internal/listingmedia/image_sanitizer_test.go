package listingmedia

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func imageRequest(body []byte) UploadRequest {
	sum := sha256.Sum256(body)
	return UploadRequest{Ordinal: 1, ContentType: "image/png", ByteSize: int64(len(body)), ChecksumSHA256: hex.EncodeToString(sum[:])}
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var b bytes.Buffer
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.NRGBA{R: 120, A: 255})
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestSanitizeImageVerifiesAndReencodesBytes(t *testing.T) {
	body := append(testPNG(t, 2, 3), []byte("synthetic-private-metadata-trailer")...)
	result, err := SanitizeImage(body, imageRequest(body))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(result.Bytes, []byte("synthetic-private-metadata")) {
		t.Fatal("private trailer survived sanitization")
	}
	if result.Width != 2 || result.Height != 3 || result.ContentType != "image/png" {
		t.Fatal("wrong sanitized image properties")
	}
	config, err := png.DecodeConfig(bytes.NewReader(result.Bytes))
	if err != nil || config.Width != 2 || config.Height != 3 {
		t.Fatal("sanitized bytes are not a valid image")
	}
	sum := sha256.Sum256(result.Bytes)
	if result.ChecksumSHA256 != hex.EncodeToString(sum[:]) {
		t.Fatal("wrong sanitized checksum")
	}
}

func TestSanitizeImageAcceptsJPEGAndWebP(t *testing.T) {
	var jpegBytes bytes.Buffer
	if err := jpeg.Encode(&jpegBytes, image.NewRGBA(image.Rect(0, 0, 2, 3)), nil); err != nil {
		t.Fatal(err)
	}
	webp, err := base64.StdEncoding.DecodeString("UklGRjQAAABXRUJQVlA4ICgAAABwAQCdASoCAAMAAUAmJaACdAGIQAD+9Yj7J3pf/43j/Jd/bvsoAAAA")
	if err != nil {
		t.Fatal(err)
	}
	for mime, body := range map[string][]byte{"image/jpeg": jpegBytes.Bytes(), "image/webp": webp} {
		r := imageRequest(body)
		r.ContentType = mime
		v, err := SanitizeImage(body, r)
		if err != nil || v.Width != 2 || v.Height != 3 || v.ContentType != "image/png" {
			t.Fatalf("%s conversion failed: %v", mime, err)
		}
	}
}

func TestSanitizeImageRejectsUntrustedInputs(t *testing.T) {
	body := testPNG(t, 2, 3)
	for _, tc := range []struct {
		name   string
		body   []byte
		mutate func(*UploadRequest)
	}{
		{"checksum", body, func(r *UploadRequest) {
			r.ChecksumSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}},
		{"size", body, func(r *UploadRequest) { r.ByteSize++ }},
		{"mime", body, func(r *UploadRequest) { r.ContentType = "image/jpeg" }},
		{"not_image", []byte("<svg onload='alert(1)'></svg>"), func(*UploadRequest) {}},
		{"excess_dimensions", testPNG(t, 8193, 1), func(*UploadRequest) {}},
		{"too_large", make([]byte, 10485761), func(*UploadRequest) {}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := imageRequest(tc.body)
			tc.mutate(&r)
			v, err := SanitizeImage(tc.body, r)
			if !errors.Is(err, ErrInvalidUpload) || len(v.Bytes) != 0 {
				t.Fatal("unsafe input produced image")
			}
		})
	}
}
