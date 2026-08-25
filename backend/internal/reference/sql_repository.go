package reference

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type sqlRepository struct{ database *sql.DB }

func NewSQLRepository(database *sql.DB) Repository {
	return sqlRepository{database: database}
}

func (r sqlRepository) Categories(ctx context.Context, locale string) ([]Category, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	rows, err := r.database.QueryContext(ctx, `
		select category.id, category.parent_id, category.slug, translation.name, category.sort_order
		from public.service_categories category
		join public.service_category_translations translation
		  on translation.category_id = category.id and translation.locale = $1
		join public.supported_locales locale
		  on locale.id = translation.locale and locale.active
		left join public.service_categories parent on parent.id = category.parent_id
		where category.active and (category.parent_id is null or parent.active)
		order by coalesce(parent.sort_order, category.sort_order),
		  case when category.parent_id is null then 0 else 1 end,
		  category.sort_order,
		  category.id
	`, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.ParentID, &category.Slug, &category.Name, &category.SortOrder); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return categories, nil
}

func (r sqlRepository) Languages(ctx context.Context, locale string) ([]Language, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	rows, err := r.database.QueryContext(ctx, `
		select language.id, translation.name, language.sort_order
		from public.spoken_languages language
		join public.spoken_language_translations translation
		  on translation.language_code = language.id and translation.locale = $1
		join public.supported_locales locale
		  on locale.id = translation.locale and locale.active
		where language.active
		order by language.sort_order, language.id
	`, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	languages := make([]Language, 0)
	for rows.Next() {
		var language Language
		if err := rows.Scan(&language.Code, &language.Name, &language.SortOrder); err != nil {
			return nil, err
		}
		languages = append(languages, language)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return languages, nil
}

func (r sqlRepository) Localities(ctx context.Context, locale string) ([]Locality, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	rows, err := r.database.QueryContext(ctx, localitySelectSQL+`
		cross join public.supported_locales locale
		where locality.active and parish.active and municipality.active and district.active
		  and locale.id = $1 and locale.active
		order by locality.name, locality.id
	`, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLocalities(rows)
}

func (r sqlRepository) NearbyLocalities(ctx context.Context, origin uuid.UUID, radiusKM int, locale string) ([]LocalityDistance, error) {
	if r.database == nil {
		return nil, errors.New("database is nil")
	}
	rows, err := r.database.QueryContext(ctx, `
		with origin as (
		  select center from public.localities where id = $1 and active
		)
		select locality.id, locality.slug, locality.name,
		  parish.name, municipality.name, district.name,
		  round(st_distance(locality.center, origin.center))::integer as distance_meters
		from origin
		cross join public.localities locality
		join public.administrative_areas parish on parish.id = locality.parent_parish_id
		join public.administrative_areas municipality on municipality.id = parish.parent_id
		join public.administrative_areas district on district.id = municipality.parent_id
		cross join public.supported_locales locale
		where locality.active and parish.active and municipality.active and district.active
		  and locale.id = $3 and locale.active
		  and st_dwithin(locality.center, origin.center, $2 * 1000)
		order by distance_meters, locality.id
	`, origin, radiusKM, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]LocalityDistance, 0)
	for rows.Next() {
		var value LocalityDistance
		if err := rows.Scan(
			&value.ID,
			&value.Slug,
			&value.Name,
			&value.ParishName,
			&value.MunicipalityName,
			&value.DistrictName,
			&value.DistanceMeters,
		); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, ErrNotFound
	}
	return values, nil
}

func (r sqlRepository) ValidateProfileReferences(ctx context.Context, references ProfileReferences) error {
	if r.database == nil {
		return errors.New("database is nil")
	}
	localityIDs := append([]uuid.UUID{references.PrimaryLocalityID}, references.ServiceLocalityIDs...)
	seenLocalities := make(map[uuid.UUID]struct{}, len(localityIDs))
	for _, id := range localityIDs {
		if id == uuid.Nil {
			return ErrInvalidReference
		}
		if _, ok := seenLocalities[id]; ok {
			continue
		}
		seenLocalities[id] = struct{}{}
		var active bool
		if err := r.database.QueryRowContext(ctx, "select active from public.localities where id = $1", id).Scan(&active); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrInvalidReference
			}
			return err
		}
		if !active {
			return ErrInvalidReference
		}
	}

	seenLanguages := make(map[string]struct{}, len(references.LanguageCodes))
	for _, code := range references.LanguageCodes {
		if _, ok := seenLanguages[code]; ok {
			continue
		}
		seenLanguages[code] = struct{}{}
		var active bool
		if err := r.database.QueryRowContext(ctx, "select active from public.spoken_languages where id = $1", code).Scan(&active); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrInvalidReference
			}
			return err
		}
		if !active {
			return ErrInvalidReference
		}
	}
	return nil
}

const localitySelectSQL = `
	select locality.id, locality.slug, locality.name,
	  parish.name, municipality.name, district.name
	from public.localities locality
	join public.administrative_areas parish on parish.id = locality.parent_parish_id
	join public.administrative_areas municipality on municipality.id = parish.parent_id
	join public.administrative_areas district on district.id = municipality.parent_id
`

func scanLocalities(rows *sql.Rows) ([]Locality, error) {
	localities := make([]Locality, 0)
	for rows.Next() {
		var locality Locality
		if err := rows.Scan(
			&locality.ID,
			&locality.Slug,
			&locality.Name,
			&locality.ParishName,
			&locality.MunicipalityName,
			&locality.DistrictName,
		); err != nil {
			return nil, err
		}
		localities = append(localities, locality)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return localities, nil
}
