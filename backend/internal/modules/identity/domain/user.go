package domain

import "time"

type UserID string
type Role string

const (
	RoleAdmin            Role = "admin"
	RoleCatalogEditor    Role = "catalog_editor"
	RoleEvidenceEditor   Role = "evidence_editor"
	RoleEvidenceReviewer Role = "evidence_reviewer"
	RolePolicyEditor     Role = "policy_editor"
	RolePolicyReviewer   Role = "policy_reviewer"
	RoleCommerceOperator Role = "commerce_operator"
	RoleContentEditor    Role = "content_editor"
	RoleAnalyst          Role = "analyst"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

type User struct {
	ID              UserID
	Email           string
	Status          UserStatus
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	Roles           []Role
}

type PasswordCredential struct {
	User         User
	PasswordHash string
}

type Principal struct {
	UserID             UserID
	Email              string
	Roles              []Role
	SessionID          string
	EmailVerifiedAt    *time.Time
	MFAAuthenticatedAt *time.Time
	MFAEnabled         bool
}

func (principal Principal) HasRole(role Role) bool {
	for _, assigned := range principal.Roles {
		if assigned == role {
			return true
		}
	}
	return false
}

type Session struct {
	ID                   string
	UserID               UserID
	TokenHash            []byte
	ExpiresAt            time.Time
	IdleExpiresAt        time.Time
	LastSeenAt           time.Time
	CreatedAt            time.Time
	RevokedAt            *time.Time
	MFAAuthenticatedAt   *time.Time
	AuthenticationMethod string
}

func (user User) CanAuthenticate() bool {
	return user.Status == UserStatusActive && user.DeletedAt == nil
}
