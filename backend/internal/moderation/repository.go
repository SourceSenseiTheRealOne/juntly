package moderation

import (
	"context"
	"errors"

	jent "github.com/SourceSenseiTheRealOne/juntly/backend/ent"
	"github.com/SourceSenseiTheRealOne/juntly/backend/ent/platformrole"
	"github.com/google/uuid"
)

type entRepository struct{ client *jent.Client }

func NewEntRepository(client *jent.Client) Repository { return entRepository{client: client} }

func (r entRepository) HasModeratorGrant(ctx context.Context, internalUserID uuid.UUID) (bool, error) {
	if r.client == nil || internalUserID == uuid.Nil {
		return false, errors.New("platform role persistence unavailable")
	}
	return r.client.PlatformRole.Query().Where(
		platformrole.InternalUserIDEQ(internalUserID),
		platformrole.RoleEQ("moderator"),
	).Exist(ctx)
}
