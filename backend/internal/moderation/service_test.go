package moderation

import (
	"context"
	"errors"
	"testing"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
	"github.com/google/uuid"
)

func TestServiceRequiresPersistedModeratorGrant(t *testing.T) {
	t.Parallel()
	moderator := users.InternalUser{ID: uuid.MustParse("eeeeeeee-eeee-deee-0eee-eeeeeeeeeeee")}
	repository := &recordingRepository{granted: true}
	resolved, err := NewService(&recordingReconciler{user: moderator}, repository).RequireModerator(context.Background(), users.VerifiedIdentity{Subject: "moderator"})
	if err != nil || resolved != moderator || repository.owner != moderator.ID {
		t.Fatalf("resolved/err/owner = %#v/%v/%s", resolved, err, repository.owner)
	}
}

func TestServiceRejectsMissingOrInvalidModeratorBeforeLifecycle(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		reconcilerErr error
		granted       bool
		want          error
	}{
		{reconcilerErr: users.ErrInvalidIdentity, want: ErrUnauthorized},
		{granted: false, want: ErrForbidden},
		{reconcilerErr: errors.New("database private detail"), want: ErrUnavailable},
	} {
		repository := &recordingRepository{granted: test.granted}
		_, err := NewService(&recordingReconciler{err: test.reconcilerErr}, repository).RequireModerator(context.Background(), users.VerifiedIdentity{})
		if !errors.Is(err, test.want) {
			t.Fatalf("error = %v, want %v", err, test.want)
		}
		if test.reconcilerErr != nil && repository.calls != 0 {
			t.Fatalf("role repository calls = %d, want 0", repository.calls)
		}
	}
}

type recordingReconciler struct {
	user users.InternalUser
	err  error
}

func (r *recordingReconciler) Reconcile(context.Context, users.VerifiedIdentity) (users.InternalUser, bool, error) {
	return r.user, false, r.err
}

type recordingRepository struct {
	owner   uuid.UUID
	granted bool
	calls   int
	err     error
}

func (r *recordingRepository) HasModeratorGrant(_ context.Context, owner uuid.UUID) (bool, error) {
	r.calls++
	r.owner = owner
	return r.granted, r.err
}
