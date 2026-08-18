package ports

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/identity/domain"
)

var ErrNotFound = errors.New("identity entity not found")
var ErrConflict = errors.New("identity entity conflicts with an existing record")
var ErrExpired = errors.New("identity credential expired")
var ErrConsumed = errors.New("identity credential already consumed")
var ErrAttemptsExceeded = errors.New("identity challenge attempts exceeded")

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

type SecurityRepository interface {
	CreateEmailVerificationToken(context.Context, string, []byte, time.Time, time.Time, domain.SecurityEvent) (string, bool, error)
	ConsumeEmailVerificationToken(context.Context, []byte, time.Time, domain.SecurityEvent) error
	CreatePasswordResetToken(context.Context, string, []byte, time.Time, time.Time, domain.SecurityEvent) (string, bool, error)
	ConsumePasswordResetToken(context.Context, []byte, string, time.Time, domain.SecurityEvent) error
	GetPasswordCredentialByID(context.Context, domain.UserID) (domain.PasswordCredential, error)
	ChangePassword(context.Context, domain.UserID, string, string, time.Time, domain.SecurityEvent) error
	ListSessions(context.Context, domain.UserID, string, time.Time) ([]domain.ActiveSession, error)
	RevokeOwnedSession(context.Context, domain.UserID, string, string, time.Time, domain.SecurityEvent) error
	RevokeOtherSessions(context.Context, domain.UserID, string, time.Time, domain.SecurityEvent) (int64, error)
	DeleteAccount(context.Context, domain.UserID, string, time.Time, domain.SecurityEvent) error
	ExportAccount(context.Context, domain.UserID, time.Time) (domain.AccountExport, error)
	RecordSecurityEvent(context.Context, domain.SecurityEvent) error
	UpsertPendingMFA(context.Context, domain.UserID, []byte, []byte, int16, time.Time, domain.SecurityEvent) (domain.MFACredential, error)
	GetMFA(context.Context, domain.UserID) (domain.MFACredential, error)
	EnableMFA(context.Context, domain.UserID, [][]byte, time.Time, domain.SecurityEvent) error
	ReplaceRecoveryCodes(context.Context, domain.UserID, [][]byte, time.Time, domain.SecurityEvent) error
	ConsumeRecoveryCode(context.Context, domain.UserID, []byte, time.Time, domain.SecurityEvent) (bool, error)
	CreateMFAChallenge(context.Context, domain.MFAChallenge, domain.SecurityEvent) error
	GetMFAChallenge(context.Context, []byte, time.Time) (domain.MFAChallenge, error)
	FailMFAChallenge(context.Context, []byte, time.Time, domain.SecurityEvent) error
	ConsumeMFAChallengeAndCreateSession(context.Context, []byte, domain.Session, time.Time, domain.SecurityEvent) (domain.User, error)
	MarkSessionMFA(context.Context, domain.UserID, string, time.Time, string, domain.SecurityEvent) error
	CleanupExpiredSecurityArtifacts(context.Context, time.Time) error
}

type VerificationMessage struct {
	Recipient string
	Token     string
	ExpiresAt time.Time
}

type PasswordResetMessage struct {
	Recipient string
	Token     string
	ExpiresAt time.Time
}

type SecurityNotification struct {
	Recipient  string
	EventType  string
	OccurredAt time.Time
}

type DeliveryReceipt struct {
	Accepted  bool
	Reference string
}

type DevelopmentMessage struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Recipient string    `json:"recipient"`
	Token     string    `json:"token,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailDelivery interface {
	SendVerification(context.Context, VerificationMessage) (DeliveryReceipt, error)
	SendPasswordReset(context.Context, PasswordResetMessage) (DeliveryReceipt, error)
	SendSecurityNotification(context.Context, SecurityNotification) (DeliveryReceipt, error)
}
