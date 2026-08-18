package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

var (
	ErrInvalidToken    = errors.New("security token is invalid")
	ErrExpiredToken    = errors.New("security token has expired")
	ErrUsedToken       = errors.New("security token was already used")
	ErrCurrentPassword = errors.New("current password is incorrect")
	ErrSessionNotFound = errors.New("session was not found")
	ErrConfirmation    = errors.New("explicit confirmation is required")
	ErrMFAUnavailable  = errors.New("multi-factor authentication is not configured")
	ErrInvalidMFACode  = errors.New("multi-factor authentication code is invalid")
	ErrMFAChallenge    = errors.New("multi-factor authentication challenge is invalid")
)

type SecretBox interface {
	Seal([]byte) (ciphertext []byte, nonce []byte, keyVersion int16, err error)
	Open([]byte, []byte, int16) ([]byte, error)
}

type TOTPVerifier interface {
	Verify(secret string, code string, now time.Time) bool
}

type SecurityConfig struct {
	VerificationTTL  time.Duration
	PasswordResetTTL time.Duration
	MFAChallengeTTL  time.Duration
	StepUpTTL        time.Duration
	SessionTTL       time.Duration
	SessionIdleTTL   time.Duration
	Issuer           string
}

type SecurityRequest struct {
	RequestID string
	Surface   string
}

type securityRequestContextKey struct{}

func WithSecurityRequest(ctx context.Context, request SecurityRequest) context.Context {
	return context.WithValue(ctx, securityRequestContextKey{}, request)
}

type RequestReceipt struct {
	Recorded          bool   `json:"recorded"`
	DeliveryAccepted  bool   `json:"delivery_accepted"`
	DeliveryReference string `json:"delivery_reference,omitempty"`
}

type MFAEnrollment struct {
	Secret          string `json:"secret"`
	ProvisioningURI string `json:"provisioning_uri"`
}

type MFAEnabled struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

type LoginChallenge struct {
	RawToken  string
	ExpiresAt time.Time
}

type SecurityService struct {
	repository ports.SecurityRepository
	passwords  PasswordHasher
	tokens     SessionTokens
	delivery   ports.EmailDelivery
	secretBox  SecretBox
	totp       TOTPVerifier
	clock      Clock
	config     SecurityConfig
}

func NewSecurityService(repository ports.SecurityRepository, passwords PasswordHasher, tokens SessionTokens,
	delivery ports.EmailDelivery, secretBox SecretBox, totp TOTPVerifier, config SecurityConfig) (*SecurityService, error) {
	if repository == nil || passwords == nil || tokens == nil || delivery == nil || secretBox == nil || totp == nil {
		return nil, errors.New("identity security dependencies are required")
	}
	if config.VerificationTTL <= 0 || config.PasswordResetTTL <= 0 || config.MFAChallengeTTL <= 0 ||
		config.StepUpTTL <= 0 || config.SessionTTL <= 0 || config.SessionIdleTTL <= 0 ||
		config.SessionIdleTTL > config.SessionTTL || strings.TrimSpace(config.Issuer) == "" {
		return nil, errors.New("identity security lifetimes or issuer are invalid")
	}
	return &SecurityService{repository: repository, passwords: passwords, tokens: tokens, delivery: delivery,
		secretBox: secretBox, totp: totp, clock: systemClock{}, config: config}, nil
}

func (service *SecurityService) RequestEmailVerification(ctx context.Context, email string) (RequestReceipt, error) {
	now := service.clock.Now()
	raw, hash, err := service.tokens.Generate()
	if err != nil {
		return RequestReceipt{}, fmt.Errorf("generate verification token: %w", err)
	}
	expires := now.Add(service.config.VerificationTTL)
	recipient, created, err := service.repository.CreateEmailVerificationToken(ctx, normalizeEmail(email), hash, expires, now,
		service.event(ctx, nil, nil, "email_verification.request", "requested", nil))
	if err != nil {
		return RequestReceipt{}, fmt.Errorf("record verification request: %w", err)
	}
	if !created {
		return RequestReceipt{Recorded: true}, nil
	}
	receipt, deliveryErr := service.delivery.SendVerification(ctx, ports.VerificationMessage{Recipient: recipient, Token: raw, ExpiresAt: expires})
	if deliveryErr != nil {
		return RequestReceipt{Recorded: true}, fmt.Errorf("deliver verification intent: %w", deliveryErr)
	}
	return RequestReceipt{Recorded: true, DeliveryAccepted: receipt.Accepted, DeliveryReference: receipt.Reference}, nil
}

func (service *SecurityService) VerifyEmail(ctx context.Context, rawToken string) error {
	hash, err := service.tokens.Hash(strings.TrimSpace(rawToken))
	if err != nil {
		return ErrInvalidToken
	}
	err = service.repository.ConsumeEmailVerificationToken(ctx, hash, service.clock.Now(),
		service.event(ctx, nil, nil, "email_verification.complete", "success", nil))
	return translateTokenError(err, "verify email")
}

func (service *SecurityService) RequestPasswordReset(ctx context.Context, email string) (RequestReceipt, error) {
	now := service.clock.Now()
	raw, hash, err := service.tokens.Generate()
	if err != nil {
		return RequestReceipt{}, fmt.Errorf("generate password reset token: %w", err)
	}
	expires := now.Add(service.config.PasswordResetTTL)
	recipient, created, err := service.repository.CreatePasswordResetToken(ctx, normalizeEmail(email), hash, expires, now,
		service.event(ctx, nil, nil, "password_reset.request", "requested", nil))
	if err != nil {
		return RequestReceipt{}, fmt.Errorf("record password reset request: %w", err)
	}
	if !created {
		return RequestReceipt{Recorded: true}, nil
	}
	receipt, deliveryErr := service.delivery.SendPasswordReset(ctx, ports.PasswordResetMessage{Recipient: recipient, Token: raw, ExpiresAt: expires})
	if deliveryErr != nil {
		return RequestReceipt{Recorded: true}, fmt.Errorf("deliver password reset intent: %w", deliveryErr)
	}
	return RequestReceipt{Recorded: true, DeliveryAccepted: receipt.Accepted, DeliveryReference: receipt.Reference}, nil
}

func (service *SecurityService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if validation := validatePassword(newPassword); validation != "" {
		return ValidationError{Fields: map[string]string{"password": validation}}
	}
	hash, err := service.tokens.Hash(strings.TrimSpace(rawToken))
	if err != nil {
		return ErrInvalidToken
	}
	passwordHash, err := service.passwords.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash reset password: %w", err)
	}
	err = service.repository.ConsumePasswordResetToken(ctx, hash, passwordHash, service.clock.Now(),
		service.event(ctx, nil, nil, "password_reset.complete", "success", nil))
	return translateTokenError(err, "reset password")
}

func (service *SecurityService) ChangePassword(ctx context.Context, principal domain.Principal, currentPassword, newPassword string) error {
	if validation := validatePassword(newPassword); validation != "" {
		return ValidationError{Fields: map[string]string{"new_password": validation}}
	}
	credential, err := service.repository.GetPasswordCredentialByID(ctx, principal.UserID)
	if err != nil {
		return fmt.Errorf("load current credential: %w", err)
	}
	valid, err := service.passwords.Verify(credential.PasswordHash, currentPassword)
	if err != nil {
		return fmt.Errorf("verify current password: %w", err)
	}
	if !valid {
		return ErrCurrentPassword
	}
	passwordHash, err := service.passwords.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hash changed password: %w", err)
	}
	now := service.clock.Now()
	if err := service.repository.ChangePassword(ctx, principal.UserID, principal.SessionID, passwordHash, now,
		service.event(ctx, &principal.UserID, &principal.SessionID, "password_change", "success", nil)); err != nil {
		return err
	}
	service.notifySecurity(ctx, principal, "password_changed", now)
	return nil
}

func (service *SecurityService) ListSessions(ctx context.Context, principal domain.Principal) ([]domain.ActiveSession, error) {
	return service.repository.ListSessions(ctx, principal.UserID, principal.SessionID, service.clock.Now())
}

func (service *SecurityService) RevokeSession(ctx context.Context, principal domain.Principal, sessionID string) error {
	err := service.repository.RevokeOwnedSession(ctx, principal.UserID, principal.SessionID, sessionID, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "session.revoke", "success", map[string]string{"target": "single"}))
	if errors.Is(err, ports.ErrNotFound) {
		return ErrSessionNotFound
	}
	return err
}

func (service *SecurityService) RevokeOtherSessions(ctx context.Context, principal domain.Principal) (int64, error) {
	return service.repository.RevokeOtherSessions(ctx, principal.UserID, principal.SessionID, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "session.revoke", "success", map[string]string{"target": "others"}))
}

func (service *SecurityService) ExportAccount(ctx context.Context, principal domain.Principal) (domain.AccountExport, error) {
	return service.repository.ExportAccount(ctx, principal.UserID, service.clock.Now())
}

func (service *SecurityService) DeleteAccount(ctx context.Context, principal domain.Principal, password, confirmation string) error {
	if err := service.repository.RecordSecurityEvent(ctx, service.event(ctx, &principal.UserID, &principal.SessionID,
		"account.delete", "requested", nil)); err != nil {
		return fmt.Errorf("record account deletion request: %w", err)
	}
	if confirmation != "DELETE" {
		return ErrConfirmation
	}
	credential, err := service.repository.GetPasswordCredentialByID(ctx, principal.UserID)
	if err != nil {
		return fmt.Errorf("load deletion credential: %w", err)
	}
	valid, err := service.passwords.Verify(credential.PasswordHash, password)
	if err != nil {
		return fmt.Errorf("verify deletion credential: %w", err)
	}
	if !valid {
		return ErrCurrentPassword
	}
	return service.repository.DeleteAccount(ctx, principal.UserID, principal.SessionID, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "account.delete", "success", nil))
}

func (service *SecurityService) BeginMFAEnrollment(ctx context.Context, principal domain.Principal, password string) (MFAEnrollment, error) {
	credential, err := service.repository.GetPasswordCredentialByID(ctx, principal.UserID)
	if err != nil {
		return MFAEnrollment{}, fmt.Errorf("load enrollment credential: %w", err)
	}
	valid, err := service.passwords.Verify(credential.PasswordHash, password)
	if err != nil {
		return MFAEnrollment{}, fmt.Errorf("verify enrollment credential: %w", err)
	}
	if !valid {
		return MFAEnrollment{}, ErrCurrentPassword
	}
	secret, err := randomTOTPSecret()
	if err != nil {
		return MFAEnrollment{}, err
	}
	ciphertext, nonce, version, err := service.secretBox.Seal([]byte(secret))
	if err != nil {
		return MFAEnrollment{}, fmt.Errorf("encrypt MFA secret: %w", err)
	}
	_, err = service.repository.UpsertPendingMFA(ctx, principal.UserID, ciphertext, nonce, version, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.enrollment_started", "requested", nil))
	if err != nil {
		return MFAEnrollment{}, err
	}
	label := url.QueryEscape(principal.Email)
	issuer := url.QueryEscape(service.config.Issuer)
	uri := "otpauth://totp/" + issuer + ":" + label + "?secret=" + secret + "&issuer=" + issuer + "&algorithm=SHA1&digits=6&period=30"
	return MFAEnrollment{Secret: secret, ProvisioningURI: uri}, nil
}

func (service *SecurityService) ConfirmMFAEnrollment(ctx context.Context, principal domain.Principal, code string) (MFAEnabled, error) {
	credential, secret, err := service.loadMFASecret(ctx, principal.UserID)
	if err != nil {
		return MFAEnabled{}, err
	}
	if credential.Status != "pending" || !service.totp.Verify(secret, code, service.clock.Now()) {
		return MFAEnabled{}, ErrInvalidMFACode
	}
	rawCodes, hashes, err := generateRecoveryCodes(10)
	if err != nil {
		return MFAEnabled{}, err
	}
	if err := service.repository.EnableMFA(ctx, principal.UserID, hashes, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.enrollment_complete", "success", nil)); err != nil {
		return MFAEnabled{}, err
	}
	if err := service.repository.MarkSessionMFA(ctx, principal.UserID, principal.SessionID, service.clock.Now(), "password_mfa",
		service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.step_up", "success", map[string]string{"source": "enrollment"})); err != nil {
		return MFAEnabled{}, err
	}
	service.notifySecurity(ctx, principal, "mfa_enabled", service.clock.Now())
	return MFAEnabled{RecoveryCodes: rawCodes}, nil
}

func (service *SecurityService) RecordAuthorizationFailure(ctx context.Context, principal domain.Principal, permission domain.Permission) error {
	return service.repository.RecordSecurityEvent(ctx, service.event(ctx, &principal.UserID, &principal.SessionID,
		"authorization", "denied", map[string]string{"permission": string(permission)}))
}

func (service *SecurityService) RegenerateRecoveryCodes(ctx context.Context, principal domain.Principal, code string) (MFAEnabled, error) {
	if err := service.VerifyStepUp(ctx, principal, code); err != nil {
		return MFAEnabled{}, err
	}
	rawCodes, hashes, err := generateRecoveryCodes(10)
	if err != nil {
		return MFAEnabled{}, err
	}
	if err := service.repository.ReplaceRecoveryCodes(ctx, principal.UserID, hashes, service.clock.Now(),
		service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.recovery_codes_regenerated", "success", nil)); err != nil {
		return MFAEnabled{}, err
	}
	service.notifySecurity(ctx, principal, "mfa_recovery_regenerated", service.clock.Now())
	return MFAEnabled{RecoveryCodes: rawCodes}, nil
}

func (service *SecurityService) notifySecurity(ctx context.Context, principal domain.Principal, eventType string, occurredAt time.Time) {
	receipt, err := service.delivery.SendSecurityNotification(ctx, ports.SecurityNotification{
		Recipient: principal.Email, EventType: eventType, OccurredAt: occurredAt,
	})
	outcome := "accepted"
	if err != nil {
		outcome = "failed"
	} else if !receipt.Accepted {
		outcome = "not_delivered"
	}
	_ = service.repository.RecordSecurityEvent(ctx, service.event(ctx, &principal.UserID, &principal.SessionID,
		"email.security_notification", outcome, map[string]string{"event_type": eventType}))
}

func (service *SecurityService) RequiresMFA(ctx context.Context, user domain.User) (bool, error) {
	credential, err := service.repository.GetMFA(ctx, user.ID)
	if errors.Is(err, ports.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return credential.Status == "enabled", nil
}

func (service *SecurityService) BeginLoginMFA(ctx context.Context, user domain.User) (LoginChallenge, error) {
	raw, hash, err := service.tokens.Generate()
	if err != nil {
		return LoginChallenge{}, err
	}
	now := service.clock.Now()
	expires := now.Add(service.config.MFAChallengeTTL)
	err = service.repository.CreateMFAChallenge(ctx, domain.MFAChallenge{UserID: user.ID, TokenHash: hash, ExpiresAt: expires, CreatedAt: now},
		service.event(ctx, &user.ID, nil, "mfa.login_challenge", "requested", nil))
	if err != nil {
		return LoginChallenge{}, err
	}
	return LoginChallenge{RawToken: raw, ExpiresAt: expires}, nil
}

func (service *SecurityService) CompleteLoginMFA(ctx context.Context, rawChallenge, code string) (AuthenticatedSession, error) {
	hash, err := service.tokens.Hash(strings.TrimSpace(rawChallenge))
	if err != nil {
		return AuthenticatedSession{}, ErrMFAChallenge
	}
	now := service.clock.Now()
	challenge, err := service.repository.GetMFAChallenge(ctx, hash, now)
	if err != nil {
		return AuthenticatedSession{}, ErrMFAChallenge
	}
	method, valid, err := service.verifyMFAOrRecovery(ctx, challenge.UserID, code, now)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	if !valid {
		_ = service.repository.FailMFAChallenge(ctx, hash, now,
			service.event(ctx, &challenge.UserID, nil, "mfa.challenge", "failure", nil))
		return AuthenticatedSession{}, ErrInvalidMFACode
	}
	session, rawSession, err := service.newAuthenticatedSession(challenge.UserID, now, method)
	if err != nil {
		return AuthenticatedSession{}, err
	}
	user, err := service.repository.ConsumeMFAChallengeAndCreateSession(ctx, hash, session, now,
		service.event(ctx, &challenge.UserID, nil, "mfa.challenge", "success", nil))
	if err != nil {
		return AuthenticatedSession{}, err
	}
	return AuthenticatedSession{User: user, RawToken: rawSession, ExpiresAt: session.ExpiresAt}, nil
}

func (service *SecurityService) VerifyStepUp(ctx context.Context, principal domain.Principal, code string) error {
	method, valid, err := service.verifyMFAOrRecovery(ctx, principal.UserID, code, service.clock.Now())
	if err != nil {
		return err
	}
	if !valid {
		_ = service.repository.RecordSecurityEvent(ctx, service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.step_up", "failure", nil))
		return ErrInvalidMFACode
	}
	return service.repository.MarkSessionMFA(ctx, principal.UserID, principal.SessionID, service.clock.Now(), method,
		service.event(ctx, &principal.UserID, &principal.SessionID, "mfa.step_up", "success", nil))
}

func (service *SecurityService) RecentMFA(principal domain.Principal) bool {
	return principal.MFAAuthenticatedAt != nil && service.clock.Now().Sub(*principal.MFAAuthenticatedAt) <= service.config.StepUpTTL
}

func (service *SecurityService) Cleanup(ctx context.Context) error {
	return service.repository.CleanupExpiredSecurityArtifacts(ctx, service.clock.Now())
}

func (service *SecurityService) verifyMFAOrRecovery(ctx context.Context, userID domain.UserID, code string, now time.Time) (string, bool, error) {
	_, secret, err := service.loadMFASecret(ctx, userID)
	if err != nil {
		return "", false, err
	}
	if service.totp.Verify(secret, code, now) {
		return "password_mfa", true, nil
	}
	hash := recoveryCodeHash(code)
	used, err := service.repository.ConsumeRecoveryCode(ctx, userID, hash, now,
		service.event(ctx, &userID, nil, "mfa.recovery_code", "success", nil))
	if err != nil {
		return "", false, err
	}
	return "password_recovery", used, nil
}

func (service *SecurityService) loadMFASecret(ctx context.Context, userID domain.UserID) (domain.MFACredential, string, error) {
	credential, err := service.repository.GetMFA(ctx, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return domain.MFACredential{}, "", ErrMFAUnavailable
	}
	if err != nil {
		return domain.MFACredential{}, "", err
	}
	plaintext, err := service.secretBox.Open(credential.SecretCiphertext, credential.SecretNonce, credential.KeyVersion)
	if err != nil {
		return domain.MFACredential{}, "", fmt.Errorf("decrypt MFA secret: %w", err)
	}
	return credential, string(plaintext), nil
}

func (service *SecurityService) newAuthenticatedSession(userID domain.UserID, now time.Time, method string) (domain.Session, string, error) {
	raw, hash, err := service.tokens.Generate()
	if err != nil {
		return domain.Session{}, "", err
	}
	expires := now.Add(service.config.SessionTTL)
	idle := now.Add(service.config.SessionIdleTTL)
	if idle.After(expires) {
		idle = expires
	}
	return domain.Session{UserID: userID, TokenHash: hash, ExpiresAt: expires, IdleExpiresAt: idle,
		LastSeenAt: now, CreatedAt: now, MFAAuthenticatedAt: &now, AuthenticationMethod: method}, raw, nil
}

func (service *SecurityService) event(ctx context.Context, userID *domain.UserID, sessionID *string,
	eventType, outcome string, metadata map[string]string) domain.SecurityEvent {
	request, _ := ctx.Value(securityRequestContextKey{}).(SecurityRequest)
	if request.Surface == "" {
		request.Surface = "api"
	}
	return domain.SecurityEvent{UserID: userID, SessionID: sessionID, Type: eventType, Outcome: outcome,
		RequestID: request.RequestID, Surface: request.Surface, Metadata: metadata, OccurredAt: service.clock.Now()}
}

func translateTokenError(err error, operation string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ports.ErrNotFound):
		return ErrInvalidToken
	case errors.Is(err, ports.ErrExpired):
		return ErrExpiredToken
	case errors.Is(err, ports.ErrConsumed):
		return ErrUsedToken
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func validatePassword(password string) string {
	characters := len([]rune(password))
	if characters < minimumPasswordCharacters {
		return "Use at least 12 characters."
	}
	if len(password) > maximumPasswordBytes {
		return "Use no more than 128 bytes."
	}
	return ""
}

func randomTOTPSecret() (string, error) {
	value := make([]byte, 20)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate MFA secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(value), nil
}

func generateRecoveryCodes(count int) ([]string, [][]byte, error) {
	raw := make([]string, 0, count)
	hashes := make([][]byte, 0, count)
	for index := 0; index < count; index++ {
		value := make([]byte, 10)
		if _, err := rand.Read(value); err != nil {
			return nil, nil, fmt.Errorf("generate recovery code: %w", err)
		}
		code := strings.ToUpper(base64.RawURLEncoding.EncodeToString(value))
		raw = append(raw, code)
		hashes = append(hashes, recoveryCodeHash(code))
	}
	return raw, hashes, nil
}

func recoveryCodeHash(code string) []byte {
	value := sha256.Sum256([]byte(strings.ToUpper(strings.TrimSpace(code))))
	return value[:]
}
