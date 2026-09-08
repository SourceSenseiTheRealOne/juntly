package listingmedia

import (
	"context"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	media "github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingmedia"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/platformrole"
	"github.com/google/uuid"
)

var _ ReadRepository = entRepository{}

func (r entRepository) ListVerified(ctx context.Context, scope ReadScope, actor, listingID uuid.UUID) ([]VerifiedMedia, error) {
	if r.client == nil {
		return nil, ErrUnavailable
	}
	if listingID == uuid.Nil {
		return nil, ErrNotFound
	}
	parent, err := r.client.Listing.Get(ctx, listingID)
	if err != nil {
		return nil, mediaReadError(err)
	}
	switch scope {
	case PublicRead:
		if parent.State != listing.StateActive {
			return nil, ErrNotFound
		}
		if err := r.requirePublicListing(ctx, parent); err != nil {
			return nil, err
		}
	case OwnerRead:
		if actor == uuid.Nil || parent.InternalUserID != actor {
			return nil, ErrNotFound
		}
	case ModeratorRead:
		if actor == uuid.Nil {
			return nil, ErrNotFound
		}
		granted, err := r.client.PlatformRole.Query().Where(platformrole.InternalUserIDEQ(actor), platformrole.RoleEQ("moderator")).Exist(ctx)
		if err != nil {
			return nil, ErrUnavailable
		}
		if !granted {
			return nil, ErrNotFound
		}
	default:
		return nil, ErrNotFound
	}
	entities, err := r.client.ListingMedia.Query().Where(media.ListingIDEQ(listingID), media.StateEQ(media.StateReady)).Order(jent.Asc(media.FieldOrdinal)).Limit(10).All(ctx)
	if err != nil {
		return nil, ErrUnavailable
	}
	result := make([]VerifiedMedia, 0, len(entities))
	for _, entity := range entities {
		// Historical ready rows without verified metadata never become visible.
		if !validVerifiedRecord(entity) {
			continue
		}
		result = append(result, VerifiedMedia{ID: entity.ID, Ordinal: entity.Ordinal, ObjectReference: *entity.VerifiedObjectReference, ByteSize: *entity.VerifiedByteSize, ChecksumSHA256: *entity.VerifiedChecksumSha256, Width: *entity.PixelWidth, Height: *entity.PixelHeight})
	}
	return result, nil
}

// Match discovery's non-localized visibility prerequisites as well as state.
func (r entRepository) requirePublicListing(ctx context.Context, parent *jent.Listing) error {
	category, err := r.client.ServiceCategory.Get(ctx, parent.CategoryID)
	if err != nil {
		return mediaReadError(err)
	}
	if !category.Active {
		return ErrNotFound
	}
	if category.ParentID != nil {
		ancestor, err := r.client.ServiceCategory.Get(ctx, *category.ParentID)
		if err != nil {
			return mediaReadError(err)
		}
		if !ancestor.Active {
			return ErrNotFound
		}
	}
	locality, err := r.client.Locality.Get(ctx, parent.PrimaryLocalityID)
	if err != nil {
		return mediaReadError(err)
	}
	if !locality.Active {
		return ErrNotFound
	}
	if _, err := r.client.ProviderProfile.Get(ctx, parent.InternalUserID); err != nil {
		return mediaReadError(err)
	}
	return nil
}

func mediaReadError(err error) error {
	if jent.IsNotFound(err) {
		return ErrNotFound
	}
	return ErrUnavailable
}
