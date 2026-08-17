package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

type repositoryStub struct {
	credential        domain.PasswordCredential
	credentialErr     error
	registeredEmail   string
	registeredHash    string
	registeredSession domain.Session
	loginSession      domain.Session
	replacementHash   *string
	principal         domain.Principal
	resolveErr        error
	revokedHash       []byte
}

func (stub *repositoryStub) RegisterWithSession(
	_ context.Context,
	email string,
	hash string,
	session domain.Session,
) (domain.User, error) {
	stub.registeredEmail = email
	stub.registeredHash = hash
	stub.registeredSession = session
	return domain.User{ID: "user-1", Email: email, Status: domain.UserStatusActive}, nil
}

func (stub *repositoryStub) GetPasswordCredentialByEmail(
	context.Context,
	string,
) (domain.PasswordCredential, error) {
	return stub.credential, stub.credentialErr
}

func (stub *repositoryStub) CreateLoginSession(
	_ context.Context,
	session domain.Session,
	_ time.Time,
	replacementHash *string,
) error {
	stub.loginSession = session
	stub.replacementHash = replacementHash
	return nil
}

func (stub *repositoryStub) ResolveSession(
	context.Context,
	[]byte,
	time.Time,
	time.Time,
) (domain.Principal, error) {
	return stub.principal, stub.resolveErr
}

func (stub *repositoryStub) RevokeSession(_ context.Context, hash []byte, _ time.Time) error {
	stub.revokedHash = hash
	return nil
}

type passwordStub struct {
	needsRehash bool
	verifyCalls int
}

func (stub *passwordStub) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (stub *passwordStub) Verify(hash, password string) (bool, error) {
	stub.verifyCalls++
	return hash == "hashed:"+password, nil
}

func (stub *passwordStub) NeedsRehash(string) bool {
	return stub.needsRehash
}

type tokenStub struct{}

func (tokenStub) Generate() (string, []byte, error) {
	return "raw-session-token", []byte("stored-token-hash"), nil
}

func (tokenStub) Hash(raw string) ([]byte, error) {
	if raw != "raw-session-token" {
		return nil, errors.New("invalid token")
	}
	return []byte("stored-token-hash"), nil
}

type clockStub struct {
	now time.Time
}

func (clock clockStub) Now() time.Time {
	return clock.now
}

func newTestService(t *testing.T, repository *repositoryStub, passwords *passwordStub) *Service {
	t.Helper()
	service, err := newService(
		repository,
		passwords,
		tokenStub{},
		clockStub{now: time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)},
		30*24*time.Hour,
		7*24*time.Hour,
	)
	if err != nil {
		t.Fatalf("newService() error = %v", err)
	}
	return service
}

func TestRegisterNormalizesCredentialsAndCreatesSession(t *testing.T) {
	repository := &repositoryStub{}
	service := newTestService(t, repository, &passwordStub{})

	session, err := service.Register(context.Background(), " Person@Example.com ", "long secure password")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.registeredEmail != "person@example.com" {
		t.Errorf("registered email = %q", repository.registeredEmail)
	}
	if repository.registeredHash == "long secure password" ||
		repository.registeredHash != "hashed:long secure password" {
		t.Error("Register() did not store only the password hash")
	}
	if session.RawToken != "raw-session-token" {
		t.Errorf("raw token = %q", session.RawToken)
	}
	if string(repository.registeredSession.TokenHash) != "stored-token-hash" {
		t.Error("Register() did not persist the session token hash")
	}
}

func TestRegisterReturnsFieldValidation(t *testing.T) {
	service := newTestService(t, &repositoryStub{}, &passwordStub{})

	_, err := service.Register(context.Background(), "invalid", "short")
	var validationError ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("Register() error = %v, want ValidationError", err)
	}
	if validationError.Fields["email"] == "" || validationError.Fields["password"] == "" {
		t.Fatalf("validation fields = %#v", validationError.Fields)
	}
}

func TestLoginUsesGenericFailureForUnknownAccount(t *testing.T) {
	repository := &repositoryStub{credentialErr: ports.ErrNotFound}
	passwords := &passwordStub{}
	service := newTestService(t, repository, passwords)

	_, err := service.Login(context.Background(), "missing@example.com", "long secure password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
	if passwords.verifyCalls != 1 {
		t.Fatalf("dummy password verification calls = %d, want 1", passwords.verifyCalls)
	}
}

func TestLoginCreatesSessionAndRehashesWhenNeeded(t *testing.T) {
	user := domain.User{ID: "user-1", Email: "person@example.com", Status: domain.UserStatusActive}
	repository := &repositoryStub{
		credential: domain.PasswordCredential{User: user, PasswordHash: "hashed:long secure password"},
	}
	passwords := &passwordStub{needsRehash: true}
	service := newTestService(t, repository, passwords)

	session, err := service.Login(context.Background(), user.Email, "long secure password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if session.User.ID != user.ID || repository.loginSession.UserID != user.ID {
		t.Fatal("Login() did not issue the session for the authenticated user")
	}
	if repository.replacementHash == nil || *repository.replacementHash != "hashed:long secure password" {
		t.Fatal("Login() did not request a password hash upgrade")
	}
}

func TestAuthenticateRejectsAnInvalidToken(t *testing.T) {
	service := newTestService(t, &repositoryStub{}, &passwordStub{})
	if _, err := service.Authenticate(context.Background(), "invalid"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Authenticate() error = %v, want ErrUnauthenticated", err)
	}
}
