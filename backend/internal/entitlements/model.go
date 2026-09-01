package entitlements

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalid      = errors.New("entitlement invalid request")
	ErrUnauthorized = errors.New("entitlement unauthorized")
	ErrForbidden    = errors.New("entitlement forbidden")
	ErrConflict     = errors.New("entitlement conflict")
	ErrUnavailable  = errors.New("entitlement unavailable")
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
	StatusExpired   Status = "expired"
)

type Plan struct {
	ID                  uuid.UUID `json:"id"`
	Slug                string    `json:"slug"`
	Name                string    `json:"name"`
	PriceMinor          int       `json:"priceMinor"`
	Currency            string    `json:"currency"`
	BillingDays         int       `json:"billingDays"`
	MaxActiveListings   int       `json:"maxActiveListings"`
	MaxPhotosPerListing int       `json:"maxPhotosPerListing"`
	AnalyticsEnabled    bool      `json:"analyticsEnabled"`
}
type PromotionPeriod struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	DurationDays int       `json:"durationDays"`
	PriceMinor   int       `json:"priceMinor"`
	Currency     string    `json:"currency"`
}
type Catalog struct {
	Plans            []Plan            `json:"plans"`
	PromotionPeriods []PromotionPeriod `json:"promotionPeriods"`
}
type Subscription struct {
	ID         uuid.UUID  `json:"id"`
	ProviderID uuid.UUID  `json:"providerId"`
	PlanID     uuid.UUID  `json:"planId"`
	Status     Status     `json:"status"`
	StartsAt   *time.Time `json:"startsAt"`
	EndsAt     *time.Time `json:"endsAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
type Promotion struct {
	ID         uuid.UUID  `json:"id"`
	ListingID  uuid.UUID  `json:"listingId"`
	ProviderID uuid.UUID  `json:"providerId"`
	PeriodID   uuid.UUID  `json:"periodId"`
	Status     Status     `json:"status"`
	StartsAt   *time.Time `json:"startsAt"`
	EndsAt     *time.Time `json:"endsAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
type Access struct {
	MaxActiveListings   int  `json:"maxActiveListings"`
	MaxPhotosPerListing int  `json:"maxPhotosPerListing"`
	AnalyticsEnabled    bool `json:"analyticsEnabled"`
}
