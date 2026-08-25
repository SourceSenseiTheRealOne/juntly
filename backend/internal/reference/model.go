package reference

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidRequest   = errors.New("invalid reference request")
	ErrInvalidReference = errors.New("invalid profile reference")
	ErrNotFound         = errors.New("reference not found")
	ErrUnavailable      = errors.New("reference data unavailable")
)

const (
	AttributionText = "© OpenStreetMap contributors"
	AttributionURL  = "https://www.openstreetmap.org/copyright"
)

type Category struct {
	ID        uuid.UUID
	ParentID  *uuid.UUID
	Slug      string
	Name      string
	SortOrder int
}

type Language struct {
	Code      string
	Name      string
	SortOrder int
}

type Locality struct {
	ID               uuid.UUID
	Slug             string
	Name             string
	ParishName       string
	MunicipalityName string
	DistrictName     string
}

type LocalityDistance struct {
	Locality
	DistanceMeters int
}

type ProfileReferences struct {
	PrimaryLocalityID  uuid.UUID
	ServiceLocalityIDs []uuid.UUID
	LanguageCodes      []string
}
