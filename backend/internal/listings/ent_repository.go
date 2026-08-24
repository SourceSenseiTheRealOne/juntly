package listings

import (
	"context"
	"errors"
	"time"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listing"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/listingevent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/locality"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/providerprofile"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/servicecategory"
	"github.com/google/uuid"
)

type entRepository struct{ client *jent.Client }

func NewEntRepository(client *jent.Client) Repository { return entRepository{client: client} }

func (r entRepository) Create(ctx context.Context, owner uuid.UUID, input CreateListing) (Listing, error) {
	if r.client == nil {
		return Listing{}, errors.New("Ent client is nil")
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return Listing{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	client := tx.Client()
	if err := validateOwnerReferences(ctx, client, owner, input); err != nil {
		return Listing{}, err
	}
	entity, err := client.Listing.Create().
		SetInternalUserID(owner).
		SetCategoryID(input.CategoryID).
		SetPrimaryLocalityID(input.PrimaryLocalityID).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetPriceType(listing.PriceType(input.PriceType)).
		SetNillablePriceMinor(input.PriceMinor).
		SetCurrency(input.Currency).
		SetTravelsToCustomer(input.TravelsToCustomer).
		SetReceivesCustomer(input.ReceivesCustomer).
		SetRemoteServices(input.RemoteServices).
		SetState(listing.StateDraft).
		SetRevision(1).
		Save(ctx)
	if err != nil {
		return Listing{}, err
	}
	if err := client.ListingEvent.Create().
		SetListingID(entity.ID).
		SetActorInternalUserID(owner).
		SetEventType(listingevent.EventTypeCreated).
		SetToState(string(StateDraft)).
		SetRevision(1).
		Exec(ctx); err != nil {
		return Listing{}, err
	}
	if err := tx.Commit(); err != nil {
		return Listing{}, err
	}
	committed = true
	return listingFromEnt(entity), nil
}

func (r entRepository) ReplaceDraft(ctx context.Context, owner, id uuid.UUID, revision int, input CreateListing) (Listing, error) {
	if r.client == nil {
		return Listing{}, errors.New("Ent client is nil")
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return Listing{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	client := tx.Client()
	if err := validateOwnerReferences(ctx, client, owner, input); err != nil {
		return Listing{}, err
	}
	affected, err := client.Listing.Update().
		Where(listing.IDEQ(id), listing.InternalUserIDEQ(owner), listing.StateEQ(listing.StateDraft), listing.RevisionEQ(revision)).
		SetCategoryID(input.CategoryID).
		SetPrimaryLocalityID(input.PrimaryLocalityID).
		SetTitle(input.Title).
		SetDescription(input.Description).
		SetPriceType(listing.PriceType(input.PriceType)).
		SetNillablePriceMinor(input.PriceMinor).
		SetTravelsToCustomer(input.TravelsToCustomer).
		SetReceivesCustomer(input.ReceivesCustomer).
		SetRemoteServices(input.RemoteServices).
		SetRevision(revision + 1).
		Save(ctx)
	if err != nil {
		return Listing{}, err
	}
	if affected != 1 {
		return Listing{}, ErrConflict
	}
	if err := client.ListingEvent.Create().
		SetListingID(id).
		SetActorInternalUserID(owner).
		SetEventType(listingevent.EventTypeUpdated).
		SetFromState(string(StateDraft)).
		SetToState(string(StateDraft)).
		SetRevision(revision + 1).
		Exec(ctx); err != nil {
		return Listing{}, err
	}
	entity, err := client.Listing.Query().Where(listing.IDEQ(id), listing.InternalUserIDEQ(owner)).Only(ctx)
	if err != nil {
		return Listing{}, err
	}
	if err := tx.Commit(); err != nil {
		return Listing{}, err
	}
	committed = true
	return listingFromEnt(entity), nil
}

func (r entRepository) FindByOwner(ctx context.Context, owner, id uuid.UUID) (*Listing, error) {
	if r.client == nil {
		return nil, errors.New("Ent client is nil")
	}
	entity, err := r.client.Listing.Query().Where(listing.IDEQ(id), listing.InternalUserIDEQ(owner)).Only(ctx)
	if jent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	value := listingFromEnt(entity)
	return &value, nil
}

func (r entRepository) ListByOwner(ctx context.Context, owner uuid.UUID) ([]Listing, error) {
	if r.client == nil {
		return nil, errors.New("Ent client is nil")
	}
	entities, err := r.client.Listing.Query().Where(listing.InternalUserIDEQ(owner)).Order(jent.Desc(listing.FieldUpdatedAt), jent.Asc(listing.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	values := make([]Listing, 0, len(entities))
	for _, entity := range entities {
		values = append(values, listingFromEnt(entity))
	}
	return values, nil
}

func validateOwnerReferences(ctx context.Context, client *jent.Client, owner uuid.UUID, input CreateListing) error {
	if _, err := client.ServiceCategory.Query().Where(servicecategory.IDEQ(input.CategoryID), servicecategory.ActiveEQ(true)).Only(ctx); err != nil {
		if jent.IsNotFound(err) {
			return ErrInvalidListing
		}
		return err
	}
	if _, err := client.Locality.Query().Where(locality.IDEQ(input.PrimaryLocalityID), locality.ActiveEQ(true)).Only(ctx); err != nil {
		if jent.IsNotFound(err) {
			return ErrInvalidListing
		}
		return err
	}
	profile, err := client.ProviderProfile.Query().Where(providerprofile.IDEQ(owner)).WithServiceLocalities().Only(ctx)
	if err != nil {
		if jent.IsNotFound(err) {
			return ErrInvalidListing
		}
		return err
	}
	if (input.TravelsToCustomer && !profile.TravelsToCustomer) || (input.ReceivesCustomer && !profile.ReceivesCustomer) || (input.RemoteServices && !profile.RemoteServices) {
		return ErrInvalidListing
	}
	for _, serviceLocality := range profile.Edges.ServiceLocalities {
		if serviceLocality.ID == input.PrimaryLocalityID {
			return nil
		}
	}
	return ErrInvalidListing
}

func listingFromEnt(entity *jent.Listing) Listing {
	var price *int
	if entity.PriceMinor != nil {
		value := *entity.PriceMinor
		price = &value
	}
	return Listing{ID: entity.ID, CreateListing: CreateListing{CategoryID: entity.CategoryID, PrimaryLocalityID: entity.PrimaryLocalityID, Title: entity.Title, Description: entity.Description, PriceType: PriceType(entity.PriceType), PriceMinor: price, Currency: entity.Currency, TravelsToCustomer: entity.TravelsToCustomer, ReceivesCustomer: entity.ReceivesCustomer, RemoteServices: entity.RemoteServices}, State: State(entity.State), Revision: entity.Revision, CreatedAt: normalizeTime(entity.CreatedAt), UpdatedAt: normalizeTime(entity.UpdatedAt)}
}

func normalizeTime(value time.Time) time.Time { return value.UTC().Truncate(time.Microsecond) }
