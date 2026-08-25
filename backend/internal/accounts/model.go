package accounts

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	CustomerEnabled       bool
	ProviderEnabled       bool
	OnboardingCompletedAt time.Time
}

type Record struct {
	InternalUserID        uuid.UUID
	ProviderEnabled       bool
	OnboardingCompletedAt time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func accountFromRecord(record Record) Account {
	return Account{
		CustomerEnabled:       true,
		ProviderEnabled:       record.ProviderEnabled,
		OnboardingCompletedAt: record.OnboardingCompletedAt,
	}
}
