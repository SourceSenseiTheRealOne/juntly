package listingmedia

import "context"

var _ VerifiedStorage = supabaseStorage{}

func (s supabaseStorage) DownloadVerified(ctx context.Context, entry VerifiedMedia) ([]byte, error) {
	if !validVerifiedMedia(entry) {
		return nil, ErrInvalidUpload
	}
	body, err := s.downloadObject(ctx, entry.ObjectReference)
	if err != nil || !matchesVerifiedBytes(body, entry) {
		return nil, ErrUnavailable
	}
	return body, nil
}
