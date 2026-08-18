package application

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

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
	User                  domain.User
	RawToken              string
	ExpiresAt             time.Time
	MFARequired           bool
	MFAChallengeToken     string
	MFAChallengeExpiresAt *time.Time
}

type LoginMFA interface {
	RequiresMFA(context.Context, domain.User) (bool, error)
	BeginLoginMFA(context.Context, domain.User) (LoginChallenge, error)
}

type SecurityAuditor interface {
	RecordSecurityEvent(context.Context, domain.SecurityEvent) error
}

type Service struct {
	repository     ports.AuthenticationRepository
	passwords      PasswordHasher
	tokens         SessionTokens
	clock          Clock
	sessionTTL     time.Duration
	sessionIdleTTL time.Duration
	dummyHash      string
	mfa            LoginMFA
	auditor        SecurityAuditor
}

func NewServiceWithMFA(
	repository ports.AuthenticationRepository,
	passwords PasswordHasher,
	tokens SessionTokens,
	mfa LoginMFA,
	sessionTTL time.Duration,
	sessionIdleTTL time.Duration,
) (*Service, error) {
	service, err := newService(repository, passwords, tokens, systemClock{}, sessionTTL, sessionIdleTTL)
	if err != nil {
		return nil, err
	}
	service.mfa = mfa
	return service, nil
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
	service := &Service{
		repository:     repository,
		passwords:      passwords,
		tokens:         tokens,
		clock:          clock,
		sessionTTL:     sessionTTL,
		sessionIdleTTL: sessionIdleTTL,
		dummyHash:      dummyHash,
	}
	if auditor, ok := repository.(SecurityAuditor); ok {
		service.auditor = auditor
	}
	return service, nil
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
	if err := service.audit(ctx, &user.ID, nil, "account.register", "success"); err != nil {
		return AuthenticatedSession{}, err
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
		if auditErr := service.audit(ctx, nil, nil, "login", "failure"); auditErr != nil {
			return AuthenticatedSession{}, auditErr
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
		if auditErr := service.audit(ctx, &credential.User.ID, nil, "login", "failure"); auditErr != nil {
			return AuthenticatedSession{}, auditErr
		}
		return AuthenticatedSession{}, ErrInvalidCredentials
	}
	if service.mfa != nil {
		required, err := service.mfa.RequiresMFA(ctx, credential.User)
		if err != nil {
			return AuthenticatedSession{}, fmt.Errorf("resolve MFA requirement: %w", err)
		}
		if required {
			challenge, err := service.mfa.BeginLoginMFA(ctx, credential.User)
			if err != nil {
				return AuthenticatedSession{}, fmt.Errorf("begin MFA login: %w", err)
			}
			return AuthenticatedSession{User: credential.User, MFARequired: true,
				MFAChallengeToken: challenge.RawToken, MFAChallengeExpiresAt: &challenge.ExpiresAt}, nil
		}
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
	if err := service.audit(ctx, &credential.User.ID, nil, "login", "success"); err != nil {
		return AuthenticatedSession{}, err
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
	return service.audit(ctx, nil, nil, "logout", "success")
}

func (service *Service) audit(ctx context.Context, userID *domain.UserID, sessionID *string, eventType, outcome string) error {
	if service.auditor == nil {
		return nil
	}
	request, _ := ctx.Value(securityRequestContextKey{}).(SecurityRequest)
	if request.Surface == "" {
		request.Surface = "api"
	}
	if err := service.auditor.RecordSecurityEvent(ctx, domain.SecurityEvent{UserID: userID, SessionID: sessionID,
		Type: eventType, Outcome: outcome, RequestID: request.RequestID, Surface: request.Surface,
		OccurredAt: service.clock.Now()}); err != nil {
		return fmt.Errorf("record authentication security event: %w", err)
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
	if message := validatePassword(password); message != "" {
		fields["password"] = message
	}
	if len(fields) > 0 {
		return "", &ValidationError{Fields: fields}
	}
	return normalizedEmail, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
