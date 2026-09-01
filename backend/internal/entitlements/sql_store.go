package entitlements

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type sqlStore struct{ database *sql.DB }

func NewSQLStore(db *sql.DB) Store { return sqlStore{database: db} }
func (s sqlStore) Catalog(ctx context.Context) (Catalog, error) {
	plans, err := s.database.QueryContext(ctx, `select id,slug,name,price_minor,currency,billing_days,max_active_listings,max_photos_per_listing,analytics_enabled from public.professional_plans where active order by price_minor,id`)
	if err != nil {
		return Catalog{}, err
	}
	defer plans.Close()
	c := Catalog{Plans: []Plan{}, PromotionPeriods: []PromotionPeriod{}}
	for plans.Next() {
		var v Plan
		if err := plans.Scan(&v.ID, &v.Slug, &v.Name, &v.PriceMinor, &v.Currency, &v.BillingDays, &v.MaxActiveListings, &v.MaxPhotosPerListing, &v.AnalyticsEnabled); err != nil {
			return Catalog{}, err
		}
		c.Plans = append(c.Plans, v)
	}
	if err = plans.Err(); err != nil {
		return Catalog{}, err
	}
	periods, err := s.database.QueryContext(ctx, `select id,slug,name,duration_days,price_minor,currency from public.promotion_periods where active order by duration_days,id`)
	if err != nil {
		return Catalog{}, err
	}
	defer periods.Close()
	for periods.Next() {
		var v PromotionPeriod
		if err := periods.Scan(&v.ID, &v.Slug, &v.Name, &v.DurationDays, &v.PriceMinor, &v.Currency); err != nil {
			return Catalog{}, err
		}
		c.PromotionPeriods = append(c.PromotionPeriods, v)
	}
	return c, periods.Err()
}
func (s sqlStore) RequestSubscription(ctx context.Context, actor, plan uuid.UUID) (Subscription, error) {
	var v Subscription
	err := s.database.QueryRowContext(ctx, `insert into public.provider_subscriptions(provider_internal_user_id,plan_id,status,starts_at,ends_at) select $1,id,case when price_minor=0 then 'active' else 'pending' end,case when price_minor=0 then timezone('utc',now()) end,case when price_minor=0 then timezone('utc',now())+billing_days*interval '1 day' end from public.professional_plans where id=$2 and active on conflict do nothing returning id,provider_internal_user_id,plan_id,status,starts_at,ends_at,created_at,updated_at`, actor, plan).Scan(&v.ID, &v.ProviderID, &v.PlanID, &v.Status, &v.StartsAt, &v.EndsAt, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Subscription{}, ErrConflict
	}
	return v, err
}
func (s sqlStore) CurrentSubscription(ctx context.Context, actor uuid.UUID) (*Subscription, error) {
	var v Subscription
	err := s.database.QueryRowContext(ctx, `select id,provider_internal_user_id,plan_id,status,starts_at,ends_at,created_at,updated_at from public.provider_subscriptions where provider_internal_user_id=$1 and status in('pending','active') order by created_at desc limit 1`, actor).Scan(&v.ID, &v.ProviderID, &v.PlanID, &v.Status, &v.StartsAt, &v.EndsAt, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}
func (s sqlStore) RequestPromotion(ctx context.Context, actor, listing, period uuid.UUID) (Promotion, error) {
	var v Promotion
	err := s.database.QueryRowContext(ctx, `insert into public.listing_promotions(listing_id,provider_internal_user_id,period_id,status,starts_at,ends_at) select l.id,$1,p.id,case when p.price_minor=0 then 'active' else 'pending' end,case when p.price_minor=0 then timezone('utc',now()) end,case when p.price_minor=0 then timezone('utc',now())+p.duration_days*interval '1 day' end from public.listings l join public.promotion_periods p on p.id=$3 and p.active where l.id=$2 and l.internal_user_id=$1 and l.state='active' on conflict do nothing returning id,listing_id,provider_internal_user_id,period_id,status,starts_at,ends_at,created_at,updated_at`, actor, listing, period).Scan(&v.ID, &v.ListingID, &v.ProviderID, &v.PeriodID, &v.Status, &v.StartsAt, &v.EndsAt, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Promotion{}, ErrConflict
	}
	return v, err
}
func (s sqlStore) ListPromotions(ctx context.Context, actor uuid.UUID) ([]Promotion, error) {
	rows, err := s.database.QueryContext(ctx, `select id,listing_id,provider_internal_user_id,period_id,status,starts_at,ends_at,created_at,updated_at from public.listing_promotions where provider_internal_user_id=$1 order by created_at desc,id`, actor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []Promotion{}
	for rows.Next() {
		var v Promotion
		if err := rows.Scan(&v.ID, &v.ListingID, &v.ProviderID, &v.PeriodID, &v.Status, &v.StartsAt, &v.EndsAt, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (s sqlStore) Access(ctx context.Context, actor uuid.UUID) (Access, error) {
	var v Access
	err := s.database.QueryRowContext(ctx, `select coalesce(p.max_active_listings,d.max_active_listings),coalesce(p.max_photos_per_listing,d.max_photos_per_listing),coalesce(p.analytics_enabled,d.analytics_enabled) from public.marketplace_entitlement_defaults d left join public.provider_subscriptions s on s.provider_internal_user_id=$1 and s.status='active' and s.ends_at>timezone('utc',now()) left join public.professional_plans p on p.id=s.plan_id where d.id=true`, actor).Scan(&v.MaxActiveListings, &v.MaxPhotosPerListing, &v.AnalyticsEnabled)
	return v, err
}
