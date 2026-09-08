package listingmedia

import (
	"context"
	"errors"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	entlistingmedia "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/provideraccess"
	"github.com/google/uuid"
)

type entRepository struct{ client *jent.Client }

func NewEntRepository(client *jent.Client) ManagedRepository { return entRepository{client: client} }
func (r entRepository) RequireEditable(ctx context.Context, owner, listingID uuid.UUID) error {
	if r.client == nil {
		return ErrUnavailable
	}
	if owner == uuid.Nil || listingID == uuid.Nil {
		return provideraccess.ErrForbidden
	}
	exists, err := r.client.Listing.Query().Where(listing.IDEQ(listingID), listing.InternalUserIDEQ(owner), listing.StateIn(listing.StateDraft, listing.StateRejected)).Exist(ctx)
	if err != nil {
		return ErrUnavailable
	}
	if !exists {
		return provideraccess.ErrForbidden
	}
	return nil
}
func (r entRepository) ReservePending(ctx context.Context, owner, listingID, mediaID uuid.UUID, request UploadRequest, objectReference string) error {
	if r.client == nil || owner == uuid.Nil || listingID == uuid.Nil || mediaID == uuid.Nil || objectReference == "" {
		return errors.New("listing media persistence unavailable")
	}
	if _, err := r.client.Listing.Query().Where(listing.IDEQ(listingID), listing.InternalUserIDEQ(owner), listing.StateIn(listing.StateDraft, listing.StateRejected)).Only(ctx); err != nil {
		return err
	}
	return r.client.ListingMedia.Create().SetID(mediaID).SetListingID(listingID).SetOrdinal(request.Ordinal).SetContentType(request.ContentType).SetByteSize(request.ByteSize).SetChecksumSha256(request.ChecksumSHA256).SetObjectReference(objectReference).SetState(entlistingmedia.StatePendingUpload).Exec(ctx)
}
