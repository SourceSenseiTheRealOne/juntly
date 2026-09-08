package listingmedia

import (
	"context"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingevent"
	media "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/google/uuid"
)

var _ FinalizationRepository = entRepository{}

func editableListing(ctx context.Context, client *jent.Client, owner, id uuid.UUID) (*jent.Listing, error) {
	if client == nil {
		return nil, ErrUnavailable
	}
	if owner == uuid.Nil || id == uuid.Nil {
		return nil, provideraccess.ErrForbidden
	}
	entity, err := client.Listing.Query().Where(listing.IDEQ(id), listing.InternalUserIDEQ(owner), listing.StateIn(listing.StateDraft, listing.StateRejected)).Only(ctx)
	if jent.IsNotFound(err) {
		return nil, provideraccess.ErrForbidden
	}
	if err != nil {
		return nil, ErrUnavailable
	}
	return entity, nil
}

func editableMedia(ctx context.Context, client *jent.Client, listingID, id uuid.UUID) (*jent.ListingMedia, error) {
	entity, err := client.ListingMedia.Query().Where(media.IDEQ(id), media.ListingIDEQ(listingID), media.StateIn(media.StatePendingUpload, media.StateReady)).Only(ctx)
	if jent.IsNotFound(err) {
		return nil, provideraccess.ErrForbidden
	}
	if err != nil {
		return nil, ErrUnavailable
	}
	return entity, nil
}

func (r entRepository) GetEditable(ctx context.Context, owner, listingID, id uuid.UUID) (MediaObject, error) {
	parent, err := editableListing(ctx, r.client, owner, listingID)
	if err != nil {
		return MediaObject{}, err
	}
	entity, err := editableMedia(ctx, r.client, listingID, id)
	if err != nil {
		return MediaObject{}, err
	}
	if entity.State == media.StateReady && !validVerifiedRecord(entity) {
		return MediaObject{}, ErrUnavailable
	}
	return MediaObject{ID: entity.ID, Request: UploadRequest{Ordinal: entity.Ordinal, ContentType: entity.ContentType, ByteSize: entity.ByteSize, ChecksumSHA256: entity.ChecksumSha256}, ObjectReference: entity.ObjectReference, Revision: parent.Revision, Ready: entity.State == media.StateReady}, nil
}

func (r entRepository) Complete(ctx context.Context, owner, listingID, id uuid.UUID, revision int, reference string, image SanitizedImage) error {
	if r.client == nil {
		return ErrUnavailable
	}
	if owner == uuid.Nil || listingID == uuid.Nil || id == uuid.Nil || revision < 1 || !validSanitizedImage(image) || reference != verifiedReference(id, image.ChecksumSHA256) {
		return ErrInvalidUpload
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return ErrUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	client := tx.Client()
	parent, err := editableListing(ctx, client, owner, listingID)
	if err != nil {
		return err
	}
	entity, err := editableMedia(ctx, client, listingID, id)
	if err != nil {
		return err
	}
	if entity.State == media.StateReady {
		if !verifiedRecordMatches(entity, reference, image) {
			return ErrInvalidUpload
		}
		return tx.Commit()
	}
	// Lock/CAS the parent before marking ready. Submit, moderation and owner
	// edits update this same revision, so none can race an unreviewed photo in.
	affected, err := client.Listing.Update().Where(listing.IDEQ(listingID), listing.InternalUserIDEQ(owner), listing.StateEQ(parent.State), listing.RevisionEQ(revision)).SetRevision(revision + 1).Save(ctx)
	if err != nil {
		return ErrUnavailable
	}
	if affected != 1 {
		return ErrConflict
	}
	affected, err = client.ListingMedia.Update().Where(media.IDEQ(id), media.ListingIDEQ(listingID), media.StateEQ(media.StatePendingUpload)).
		SetState(media.StateReady).SetVerifiedObjectReference(reference).SetVerifiedChecksumSha256(image.ChecksumSHA256).
		SetVerifiedByteSize(int64(len(image.Bytes))).SetPixelWidth(image.Width).SetPixelHeight(image.Height).Save(ctx)
	if err != nil {
		return ErrUnavailable
	}
	if affected != 1 {
		return ErrConflict
	}
	if err := client.ListingEvent.Create().SetListingID(listingID).SetActorInternalUserID(owner).
		SetEventType(listingevent.EventTypeUpdated).SetFromState(string(parent.State)).SetToState(string(parent.State)).SetRevision(revision + 1).Exec(ctx); err != nil {
		return ErrUnavailable
	}
	return tx.Commit()
}

func verifiedReference(id uuid.UUID, checksum string) string {
	return "verified/" + id.String() + "/" + checksum + ".png"
}

func validVerifiedRecord(entity *jent.ListingMedia) bool {
	if entity.VerifiedObjectReference == nil || entity.VerifiedChecksumSha256 == nil || entity.VerifiedByteSize == nil || entity.PixelWidth == nil || entity.PixelHeight == nil {
		return false
	}
	request := UploadRequest{Ordinal: entity.Ordinal, ContentType: "image/png", ByteSize: *entity.VerifiedByteSize, ChecksumSHA256: *entity.VerifiedChecksumSha256}
	return validUploadRequest(request) && *entity.VerifiedObjectReference == verifiedReference(entity.ID, *entity.VerifiedChecksumSha256) && *entity.PixelWidth > 0 && *entity.PixelHeight > 0 && *entity.PixelWidth <= 8192 && *entity.PixelHeight <= 8192 && int64(*entity.PixelWidth)*int64(*entity.PixelHeight) <= 16_000_000
}

func verifiedRecordMatches(entity *jent.ListingMedia, reference string, image SanitizedImage) bool {
	return validVerifiedRecord(entity) && *entity.VerifiedObjectReference == reference && *entity.VerifiedChecksumSha256 == image.ChecksumSHA256 && *entity.VerifiedByteSize == int64(len(image.Bytes)) && *entity.PixelWidth == image.Width && *entity.PixelHeight == image.Height
}
