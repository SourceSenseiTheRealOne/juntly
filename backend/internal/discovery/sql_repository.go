package discovery

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type sqlRepository struct{ database *sql.DB }

func NewSQLRepository(database *sql.DB) Repository { return sqlRepository{database: database} }

func (r sqlRepository) Search(ctx context.Context, request Request) ([]Listing, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	rows, err := r.database.QueryContext(ctx, publicListingSearchSQL,
		request.Locale,
		nilUUID(request.CategoryID),
		request.Query,
		nilUUID(request.NearLocalityID),
		request.RadiusKM,
		nilPriceType(request.PriceType),
		nilServiceMode(request.ServiceMode),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanListings(rows)
}

func (r sqlRepository) Get(ctx context.Context, id uuid.UUID, locale string) (*Listing, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	row := r.database.QueryRowContext(ctx, publicListingDetailSQL, locale, id)
	value, err := scanListing(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func nilUUID(value uuid.UUID) any {
	if value == uuid.Nil {
		return nil
	}
	return value
}

func nilPriceType(value PriceType) any {
	if value == "" {
		return nil
	}
	return string(value)
}

func nilServiceMode(value ServiceMode) any {
	if value == "" {
		return nil
	}
	return string(value)
}

const publicListingColumns = `
  listing.id, listing.title, listing.description,
  category.id, category.slug, category_translation.name,
  locality.id, locality.slug, locality.name,
  listing.price_type, listing.price_minor, listing.currency,
  listing.travels_to_customer, listing.receives_customer, listing.remote_services,
  profile.display_name, profile.provider_type,
  exists(select 1 from public.listing_promotions promotion where promotion.listing_id=listing.id and promotion.status='active' and promotion.starts_at<=timezone('utc',now()) and promotion.ends_at>timezone('utc',now())) as promoted,
  listing.updated_at
`

const publicListingFrom = `
from public.listings listing
join public.provider_profiles profile on profile.internal_user_id = listing.internal_user_id
join public.service_categories category on category.id = listing.category_id and category.active
left join public.service_categories parent on parent.id = category.parent_id
join public.service_category_translations category_translation
  on category_translation.category_id = category.id and category_translation.locale = $1
join public.supported_locales locale on locale.id = category_translation.locale and locale.active
join public.localities locality on locality.id = listing.primary_locality_id and locality.active
where listing.state = 'active'
  and (category.parent_id is null or parent.active)
`

const publicListingSearchSQL = `
select` + publicListingColumns + publicListingFrom + `
  and ($2::uuid is null or listing.category_id = $2)
  and ($3::text = '' or listing.title ilike '%' || $3 || '%' or listing.description ilike '%' || $3 || '%')
  and ($4::uuid is null or exists (select 1 from public.localities origin where origin.id = $4 and origin.active))
  and ($4::uuid is null or st_dwithin(
    locality.center,
    (select origin.center from public.localities origin where origin.id = $4 and origin.active),
    $5 * 1000
  ))
  and ($6::text is null or listing.price_type = $6)
  and ($7::text is null or
    ($7 = 'travels_to_customer' and listing.travels_to_customer) or
    ($7 = 'receives_customer' and listing.receives_customer) or
    ($7 = 'remote_services' and listing.remote_services)
  )
order by
  promoted desc,
  case when $2::uuid is null then 1 else 0 end,
  case when $4::uuid is null then 0 else round(st_distance(
    locality.center,
    (select origin.center from public.localities origin where origin.id = $4 and origin.active)
  ))::integer end,
  listing.updated_at desc,
  listing.id
`

const publicListingDetailSQL = `
select` + publicListingColumns + publicListingFrom + `
  and listing.id = $2
`

type rowScanner interface{ Scan(...any) error }

func scanListings(rows *sql.Rows) ([]Listing, error) {
	values := make([]Listing, 0)
	for rows.Next() {
		value, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func scanListing(row rowScanner) (Listing, error) {
	var value Listing
	var priceMinor sql.NullInt64
	if err := row.Scan(
		&value.ID,
		&value.Title,
		&value.Description,
		&value.CategoryID,
		&value.CategorySlug,
		&value.CategoryName,
		&value.PrimaryLocalityID,
		&value.LocalitySlug,
		&value.LocalityName,
		&value.PriceType,
		&priceMinor,
		&value.Currency,
		&value.TravelsToCustomer,
		&value.ReceivesCustomer,
		&value.RemoteServices,
		&value.ProviderDisplayName,
		&value.ProviderType,
		&value.Promoted,
		&value.UpdatedAt,
	); err != nil {
		return Listing{}, err
	}
	if priceMinor.Valid {
		minor := int(priceMinor.Int64)
		value.PriceMinor = &minor
	}
	return value, nil
}
