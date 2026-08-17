package application

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

const (
	minimumPasswordCharacters = 12
	maximumPasswordBytes      = 128
	maximumEmailBytes         = 254
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyUsed   = errors.New("email is already registered")
	ErrUnauthenticated    = errors.New("authentication required")
)

type ValidationError struct {
	Fields map[string]string
}

func (err ValidationError) Error() string {
	return "authentication input is invalid"
}

type PasswordHasher interface {
	Hash(string) (string, error)
	Verify(string, string) (bool, error)
	NeedsRehash(string) bool
}

type SessionTokens interface {
	Generate() (string, []byte, error)
	Hash(string) ([]byte, error)
}

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type AuthenticatedSession struct {
	User      domain.User
	RawToken  string
	ExpiresAt time.Time
}

type Service struct {
	repository     ports.AuthenticationRepository
	passwords      PasswordHasher
	tokens         SessionTokens
	clock          Clock
	sessionTTL     time.Duration
	sessionIdleTTL time.Duration
	dummyHash      string
}

func NewService(
	repository ports.AuthenticationRepository,
	passwords PasswordHasher,
	tokens SessionTokens,
	sessionTTL time.Duration,
	sessionIdleTTL time.Duration,
) (*Service, error) {
	return newService(repository, passwords, tokens, systemClock{}, sessionTTL, sessionIdleTTL)
}

func newService(
	repository ports.AuthenticationRepository,
	passwords PasswordHasher,
	tokens SessionTokens,
	clock Clock,
	sessionTTL time.Duration,
	sessionIdleTTL time.Duration,
) (*Service, error) {
	if sessionTTL <= 0 || sessionIdleTTL <= 0 || sessionIdleTTL > sessionTTL {
		return nil, errors.New("invalid session lifetimes")
	}
	dummyHash, err := passwords.Hash("not-a-real-account-password")
	if err != nil {
		return nil, fmt.Errorf("create login timing hash: %w", err)
	}
	return &Service{
		repository:     repository,
		passwords:      passwords,
		tokens:         tokens,
		clock:          clock,
		sessionTTL:     sessionTTL,
		sessionIdleTTL: sessionIdleTTL,
		dummyHash:      dummyHash,
	}, nil
}

func (service *Service) Register(
	ctx context.Context,
	email string,
	password string,
) (AuthenticatedSession, error) {
	normalizedEmail, validationErr := validateCredentials(email, password)
	if validationErr != nil {
		return AuthenticatedSession{}, *validationErr
	}
	passwordHash, err := service.passwords.Hash(password)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("hash password: %w", err)
	}
	session, rawToken, err := service.newSession("")
	if err != nil {
		return AuthenticatedSession{}, err
	}
	user, err := service.repository.RegisterWithSession(ctx, normalizedEmail, passwordHash, session)
	if errors.Is(err, ports.ErrConflict) {
		return AuthenticatedSession{}, ErrEmailAlreadyUsed
	}
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("register account: %w", err)
	}
	return AuthenticatedSession{User: user, RawToken: rawToken, ExpiresAt: session.ExpiresAt}, nil
}

func (service *Service) Login(
	ctx context.Context,
	email string,
	password string,
) (AuthenticatedSession, error) {
	normalizedEmail := normalizeEmail(email)
	credential, err := service.repository.GetPasswordCredentialByEmail(ctx, normalizedEmail)
	if errors.Is(err, ports.ErrNotFound) {
		if _, verifyErr := service.passwords.Verify(service.dummyHash, password); verifyErr != nil {
			return AuthenticatedSession{}, fmt.Errorf("verify timing hash: %w", verifyErr)
		}
		return AuthenticatedSession{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("load credential: %w", err)
	}
	valid, err := service.passwords.Verify(credential.PasswordHash, password)
	if err != nil {
		return AuthenticatedSession{}, fmt.Errorf("verify password: %w", err)
	}
	if !valid || !credential.User.CanAuthenticate() {
		return AuthenticatedSession{}, ErrInvalidCredentials
	}

	var replacementHash *string
	if service.passwords.NeedsRehash(credential.PasswordHash) {
		hash, err := service.passwords.Hash(password)
		if err != nil {
			return AuthenticatedSession{}, fmt.Errorf("rehash password: %w", err)
		}
		replacementHash = &hash
	}
	session, rawToken, err := service.newSession(credential.User.ID)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	if err := service.repository.CreateLoginSession(
		ctx,
		session,
		service.clock.Now(),
		replacementHash,
	); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return AuthenticatedSession{}, ErrInvalidCredentials
		}
		return AuthenticatedSession{}, fmt.Errorf("create login session: %w", err)
	}
	return AuthenticatedSession{
		User:      credential.User,
		RawToken:  rawToken,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (service *Service) Authenticate(ctx context.Context, rawToken string) (domain.Principal, error) {
	tokenHash, err := service.tokens.Hash(rawToken)
	if err != nil {
		return domain.Principal{}, ErrUnauthenticated
	}
	now := service.clock.Now()
	principal, err := service.repository.ResolveSession(
		ctx,
		tokenHash,
		now,
		now.Add(service.sessionIdleTTL),
	)
	if errors.Is(err, ports.ErrNotFound) {
		return domain.Principal{}, ErrUnauthenticated
	}
	if err != nil {
		return domain.Principal{}, fmt.Errorf("resolve authenticated session: %w", err)
	}
	return principal, nil
}

func (service *Service) Logout(ctx context.Context, rawToken string) error {
	tokenHash, err := service.tokens.Hash(rawToken)
	if err != nil {
		return nil
	}
	if err := service.repository.RevokeSession(ctx, tokenHash, service.clock.Now()); err != nil {
		return fmt.Errorf("revoke authenticated session: %w", err)
	}
	return nil
}

func (service *Service) newSession(userID domain.UserID) (domain.Session, string, error) {
	rawToken, tokenHash, err := service.tokens.Generate()
	if err != nil {
		return domain.Session{}, "", fmt.Errorf("generate session: %w", err)
	}
	now := service.clock.Now()
	expiresAt := now.Add(service.sessionTTL)
	idleExpiresAt := now.Add(service.sessionIdleTTL)
	if idleExpiresAt.After(expiresAt) {
		idleExpiresAt = expiresAt
	}
	return domain.Session{
		UserID:        userID,
		TokenHash:     tokenHash,
		ExpiresAt:     expiresAt,
		IdleExpiresAt: idleExpiresAt,
		LastSeenAt:    now,
		CreatedAt:     now,
	}, rawToken, nil
}

func validateCredentials(email, password string) (string, *ValidationError) {
	fields := make(map[string]string)
	normalizedEmail := normalizeEmail(email)
	parsed, err := mail.ParseAddress(normalizedEmail)
	if err != nil || parsed.Address != normalizedEmail || len(normalizedEmail) > maximumEmailBytes {
		fields["email"] = "Enter a valid email address."
	}
	passwordCharacters := utf8.RuneCountInString(password)
	if passwordCharacters < minimumPasswordCharacters {
		fields["password"] = "Use at least 12 characters."
	} else if len(password) > maximumPasswordBytes {
		fields["password"] = "Use no more than 128 bytes."
	}
	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	return normalizedEmail, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
