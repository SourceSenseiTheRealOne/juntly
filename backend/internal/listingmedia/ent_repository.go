package listingmedia

import (
	"context"
	"errors"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	entlistingmedia "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/google/uuid"
)

type entRepository struct{ client *jent.Client }

func NewEntRepository(client *jent.Client) Repository { return entRepository{client: client} }
func (r entRepository) ReservePending(ctx context.Context, owner, listingID, mediaID uuid.UUID, request UploadRequest, objectReference string) error {
	if r.client == nil || owner == uuid.Nil || listingID == uuid.Nil || mediaID == uuid.Nil || objectReference == "" {
		return errors.New("listing media persistence unavailable")
	}
	if _, err := r.client.Listing.Query().Where(listing.IDEQ(listingID), listing.InternalUserIDEQ(owner), listing.StateIn(listing.StateDraft, listing.StateRejected)).Only(ctx); err != nil {
		return err
	}
	return r.client.ListingMedia.Create().SetID(mediaID).SetListingID(listingID).SetOrdinal(request.Ordinal).SetContentType(request.ContentType).SetByteSize(request.ByteSize).SetChecksumSha256(request.ChecksumSHA256).SetObjectReference(objectReference).SetState(entlistingmedia.StatePendingUpload).Exec(ctx)
}
