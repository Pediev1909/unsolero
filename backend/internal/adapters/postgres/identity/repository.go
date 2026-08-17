package identitypostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repository *Repository) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	user, err := repository.getUser(ctx, `
		SELECT id, email, status, email_verified_at, last_login_at,
			created_at, updated_at, deleted_at
		FROM identity.users
		WHERE id = $1`, string(id))
	if err != nil {
		return domain.User{}, err
	}
	if err := repository.loadRoles(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (repository *Repository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := repository.getUser(ctx, `
		SELECT id, email, status, email_verified_at, last_login_at,
			created_at, updated_at, deleted_at
		FROM identity.users
		WHERE email = $1`, normalizeEmail(email))
	if err != nil {
		return domain.User{}, err
	}
	if err := repository.loadRoles(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (repository *Repository) GetPasswordCredentialByEmail(
	ctx context.Context,
	email string,
) (domain.PasswordCredential, error) {
	var credential domain.PasswordCredential
	var emailVerifiedAt sql.NullTime
	var lastLoginAt sql.NullTime
	var deletedAt sql.NullTime
	var passwordHash sql.NullString
	err := repository.pool.QueryRow(ctx, `
		SELECT id, email, status, email_verified_at, last_login_at,
			created_at, updated_at, deleted_at, password_hash
		FROM identity.users
		WHERE email = $1`, normalizeEmail(email)).Scan(
		&credential.User.ID,
		&credential.User.Email,
		&credential.User.Status,
		&emailVerifiedAt,
		&lastLoginAt,
		&credential.User.CreatedAt,
		&credential.User.UpdatedAt,
		&deletedAt,
		&passwordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) || !passwordHash.Valid {
		return domain.PasswordCredential{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.PasswordCredential{}, fmt.Errorf("get password credential: %w", err)
	}
	credential.PasswordHash = passwordHash.String
	assignOptionalUserTimes(&credential.User, emailVerifiedAt, lastLoginAt, deletedAt)
	if err := repository.loadRoles(ctx, &credential.User); err != nil {
		return domain.PasswordCredential{}, err
	}
	return credential, nil
}

func (repository *Repository) RegisterWithSession(
	ctx context.Context,
	email string,
	passwordHash string,
	session domain.Session,
) (domain.User, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("begin registration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var user domain.User
	err = tx.QueryRow(ctx, `
		INSERT INTO identity.users (email, password_hash, status)
		VALUES ($1, $2, 'active')
		RETURNING id, email, status, created_at, updated_at`,
		normalizeEmail(email), passwordHash,
	).Scan(&user.ID, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, ports.ErrConflict
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	if err := insertSession(ctx, tx, user.ID, session); err != nil {
		return domain.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.User{}, fmt.Errorf("commit registration: %w", err)
	}
	return user, nil
}

func (repository *Repository) CreateLoginSession(
	ctx context.Context,
	session domain.Session,
	loggedInAt time.Time,
	replacementPasswordHash *string,
) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin login: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE identity.users
		SET last_login_at = $2,
			password_hash = COALESCE($3, password_hash),
			updated_at = CASE WHEN $3::text IS NULL THEN updated_at ELSE $2 END
		WHERE id = $1 AND status = 'active' AND deleted_at IS NULL`,
		session.UserID, loggedInAt, replacementPasswordHash,
	)
	if err != nil {
		return fmt.Errorf("record login: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	if err := insertSession(ctx, tx, session.UserID, session); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit login: %w", err)
	}
	return nil
}

func (repository *Repository) ResolveSession(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
	nextIdleExpiration time.Time,
) (domain.Principal, error) {
	var principal domain.Principal
	err := repository.pool.QueryRow(ctx, `
		UPDATE identity.sessions AS sessions
		SET last_seen_at = $2,
			idle_expires_at = LEAST(sessions.expires_at, $3)
		FROM identity.users AS users
		WHERE sessions.user_id = users.id
			AND sessions.token_hash = $1
			AND sessions.revoked_at IS NULL
			AND sessions.expires_at > $2
			AND sessions.idle_expires_at > $2
			AND users.status = 'active'
			AND users.deleted_at IS NULL
		RETURNING users.id, users.email,
			ARRAY(SELECT role_key FROM identity.user_roles WHERE user_id = users.id ORDER BY role_key)`,
		tokenHash, now, nextIdleExpiration).Scan(
		&principal.UserID,
		&principal.Email,
		&principal.Roles,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Principal{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.Principal{}, fmt.Errorf("resolve session: %w", err)
	}
	return principal, nil
}

func (repository *Repository) loadRoles(ctx context.Context, user *domain.User) error {
	var roles []domain.Role
	if err := repository.pool.QueryRow(ctx, `
		SELECT ARRAY(SELECT role_key FROM identity.user_roles WHERE user_id = $1 ORDER BY role_key)`,
		user.ID,
	).Scan(&roles); err != nil {
		return fmt.Errorf("load user roles: %w", err)
	}
	user.Roles = roles
	return nil
}

func (repository *Repository) RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	if _, err := repository.pool.Exec(ctx, `
		UPDATE identity.sessions
		SET revoked_at = COALESCE(revoked_at, $2)
		WHERE token_hash = $1`, tokenHash, now); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (repository *Repository) getUser(
	ctx context.Context,
	query string,
	argument string,
) (domain.User, error) {
	var user domain.User
	var emailVerifiedAt sql.NullTime
	var lastLoginAt sql.NullTime
	var deletedAt sql.NullTime
	err := repository.pool.QueryRow(ctx, query, argument).Scan(
		&user.ID,
		&user.Email,
		&user.Status,
		&emailVerifiedAt,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	assignOptionalUserTimes(&user, emailVerifiedAt, lastLoginAt, deletedAt)
	return user, nil
}

func insertSession(
	ctx context.Context,
	tx pgx.Tx,
	userID domain.UserID,
	session domain.Session,
) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO identity.sessions (
			user_id, token_hash, expires_at, idle_expires_at, last_seen_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		userID,
		session.TokenHash,
		session.ExpiresAt,
		session.IdleExpiresAt,
		session.LastSeenAt,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func assignOptionalUserTimes(
	user *domain.User,
	emailVerifiedAt sql.NullTime,
	lastLoginAt sql.NullTime,
	deletedAt sql.NullTime,
) {
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = &emailVerifiedAt.Time
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

var _ ports.UserRepository = (*Repository)(nil)
var _ ports.AuthenticationRepository = (*Repository)(nil)
