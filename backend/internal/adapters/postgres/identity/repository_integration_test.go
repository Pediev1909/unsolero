package identitypostgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

func TestAuthenticationSessionLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	repository := New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := fmt.Sprintf("authentication-%d@example.invalid", now.UnixNano())
	tokenDigest := sha256.Sum256([]byte(email))
	tokenHash := tokenDigest[:]
	session := domain.Session{
		TokenHash:     tokenHash,
		ExpiresAt:     now.Add(24 * time.Hour),
		IdleExpiresAt: now.Add(time.Hour),
		LastSeenAt:    now,
		CreatedAt:     now,
	}
	user, err := repository.RegisterWithSession(ctx, email, "argon2id-test-hash", session)
	if err != nil {
		t.Fatalf("RegisterWithSession() error = %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM identity.users WHERE id = $1", user.ID)
	})
	if _, err := pool.Exec(ctx, `INSERT INTO identity.user_roles (user_id, role_key) VALUES ($1, 'admin')`, user.ID); err != nil {
		t.Fatalf("grant admin role: %v", err)
	}

	credential, err := repository.GetPasswordCredentialByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetPasswordCredentialByEmail() error = %v", err)
	}
	if credential.PasswordHash != "argon2id-test-hash" {
		t.Fatal("password credential was not stored as supplied hash data")
	}
	principal, err := repository.ResolveSession(ctx, tokenHash, now.Add(time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	if principal.UserID != user.ID || principal.Email != email {
		t.Fatalf("resolved principal = %#v", principal)
	}
	if !principal.HasRole(domain.RoleAdmin) {
		t.Fatalf("resolved principal roles = %#v", principal.Roles)
	}
	if err := repository.RevokeSession(ctx, tokenHash, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	_, err = repository.ResolveSession(ctx, tokenHash, now.Add(3*time.Minute), now.Add(time.Hour))
	if !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("ResolveSession() after revoke error = %v, want ErrNotFound", err)
	}
}
