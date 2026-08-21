package authn

import (
	"context"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

type Verifier interface {
	Verify(context.Context, string) (users.VerifiedIdentity, error)
}
