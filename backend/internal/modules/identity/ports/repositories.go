package ports

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/identity/domain"
)

var ErrNotFound = errors.New("identity entity not found")
var ErrConflict = errors.New("identity entity conflicts with an existing record")

type UserRepository interface {
	GetByID(context.Context, domain.UserID) (domain.User, error)
	GetByEmail(context.Context, string) (domain.User, error)
}

type AuthenticationRepository interface {
	RegisterWithSession(
		context.Context,
		string,
		string,
		domain.Session,
	) (domain.User, error)
	GetPasswordCredentialByEmail(context.Context, string) (domain.PasswordCredential, error)
	CreateLoginSession(context.Context, domain.Session, time.Time, *string) error
	ResolveSession(context.Context, []byte, time.Time, time.Time) (domain.Principal, error)
	RevokeSession(context.Context, []byte, time.Time) error
}
