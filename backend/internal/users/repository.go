package users

import (
	"context"
	"errors"
)

var (
	ErrSubjectConflict = errors.New("internal user subject conflict")
	ErrUnavailable     = errors.New("internal user persistence unavailable")
)

type Repository interface {
	FindBySubject(context.Context, string) (InternalUser, bool, error)
	Create(context.Context, string) (InternalUser, error)
}
