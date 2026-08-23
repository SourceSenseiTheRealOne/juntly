package providers

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile = errors.New("invalid provider profile")
	ErrUnavailable    = errors.New("provider profile unavailable")
)

type ProviderType string

const (
	ProviderTypeIndividual   ProviderType = "individual"
	ProviderTypeProfessional ProviderType = "professional"
	ProviderTypeBusiness     ProviderType = "business"
)

type ReplaceProfile struct {
	DisplayName         string
	ProviderType        ProviderType
	Bio                 string
	PrimaryLocalityID   uuid.UUID
	ServiceLocalityIDs  []uuid.UUID
	MaxTravelDistanceKM int
	TravelsToCustomer   bool
	ReceivesCustomer    bool
	RemoteServices      bool
	LanguageCodes       []string
}

type Profile struct {
	DisplayName         string
	ProviderType        ProviderType
	Bio                 string
	PrimaryLocalityID   uuid.UUID
	ServiceLocalityIDs  []uuid.UUID
	MaxTravelDistanceKM int
	TravelsToCustomer   bool
	ReceivesCustomer    bool
	RemoteServices      bool
	LanguageCodes       []string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
