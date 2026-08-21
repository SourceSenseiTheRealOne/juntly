package users

import (
	"time"

	"github.com/google/uuid"
)

const maxSubjectLength = 255

type VerifiedIdentity struct {
	Subject string
}

type InternalUser struct {
	ID        uuid.UUID
	CreatedAt time.Time
}
