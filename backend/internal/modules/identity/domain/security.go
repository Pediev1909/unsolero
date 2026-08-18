package domain

import "time"

type Permission string

const (
	PermissionAdminRead          Permission = "admin.read"
	PermissionAnalyticsRead      Permission = "analytics.read"
	PermissionAnalyticsExport    Permission = "analytics.export"
	PermissionAnalyticsRawRead   Permission = "analytics.raw.read"
	PermissionConsentRead        Permission = "analytics.consent.read"
	PermissionCatalogRead        Permission = "catalog.read"
	PermissionCatalogCreate      Permission = "catalog.create"
	PermissionCatalogUpdate      Permission = "catalog.update"
	PermissionCatalogDelete      Permission = "catalog.delete"
	PermissionEvidenceRead       Permission = "evidence.read"
	PermissionEvidenceCreate     Permission = "evidence.create"
	PermissionEvidenceApprove    Permission = "evidence.approve"
	PermissionEvidencePublish    Permission = "evidence.publish"
	PermissionPolicyRead         Permission = "policy.read"
	PermissionPolicyCreate       Permission = "policy.create"
	PermissionPolicyApprove      Permission = "policy.approve"
	PermissionPolicyActivate     Permission = "policy.activate"
	PermissionCommerceRead       Permission = "commerce.read"
	PermissionCommerceCreate     Permission = "commerce.create"
	PermissionCommerceUpdate     Permission = "commerce.update"
	PermissionCommerceActivate   Permission = "commerce.activate"
	PermissionContentRead        Permission = "content.read"
	PermissionContentCreate      Permission = "content.create"
	PermissionContentUpdate      Permission = "content.update"
	PermissionContentDelete      Permission = "content.delete"
	PermissionUsersRead          Permission = "users.read"
	PermissionSecurityEventsRead Permission = "security_events.read"
)

var rolePermissions = map[Role]map[Permission]struct{}{
	RoleCatalogEditor: permissions(
		PermissionAdminRead, PermissionCatalogRead, PermissionCatalogCreate,
		PermissionCatalogUpdate, PermissionEvidenceRead,
	),
	RoleEvidenceEditor: permissions(
		PermissionAdminRead, PermissionCatalogRead, PermissionEvidenceRead,
		PermissionEvidenceCreate,
	),
	RoleEvidenceReviewer: permissions(
		PermissionAdminRead, PermissionCatalogRead, PermissionEvidenceRead,
		PermissionEvidenceApprove, PermissionEvidencePublish,
	),
	RolePolicyEditor: permissions(
		PermissionAdminRead, PermissionPolicyRead, PermissionPolicyCreate,
	),
	RolePolicyReviewer: permissions(
		PermissionAdminRead, PermissionPolicyRead, PermissionPolicyApprove,
		PermissionPolicyActivate,
	),
	RoleCommerceOperator: permissions(
		PermissionAdminRead, PermissionCommerceRead, PermissionCommerceCreate,
		PermissionCommerceUpdate, PermissionCommerceActivate,
	),
	RoleContentEditor: permissions(
		PermissionAdminRead, PermissionContentRead, PermissionContentCreate,
		PermissionContentUpdate, PermissionContentDelete,
	),
	RoleAnalyst: permissions(
		PermissionAdminRead, PermissionAnalyticsRead, PermissionAnalyticsExport,
		PermissionCatalogRead, PermissionCommerceRead,
	),
}

func permissions(values ...Permission) map[Permission]struct{} {
	result := make(map[Permission]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func (principal Principal) HasPermission(permission Permission) bool {
	if principal.HasRole(RoleAdmin) {
		return true
	}
	for _, role := range principal.Roles {
		if _, ok := rolePermissions[role][permission]; ok {
			return true
		}
	}
	return false
}

func (principal Principal) IsPrivileged() bool {
	return len(principal.Roles) > 0
}

type SecurityEvent struct {
	ID         string
	UserID     *UserID
	SessionID  *string
	Type       string
	Outcome    string
	RequestID  string
	Surface    string
	Metadata   map[string]string
	OccurredAt time.Time
}

type ActiveSession struct {
	ID                   string     `json:"id"`
	Current              bool       `json:"current"`
	CreatedAt            time.Time  `json:"created_at"`
	LastSeenAt           time.Time  `json:"last_seen_at"`
	ExpiresAt            time.Time  `json:"expires_at"`
	IdleExpiresAt        time.Time  `json:"idle_expires_at"`
	MFAAuthenticatedAt   *time.Time `json:"mfa_authenticated_at,omitempty"`
	AuthenticationMethod string     `json:"authentication_method"`
}

type MFACredential struct {
	ID               string
	UserID           UserID
	SecretCiphertext []byte
	SecretNonce      []byte
	KeyVersion       int16
	Status           string
	CreatedAt        time.Time
	VerifiedAt       *time.Time
}

type MFAChallenge struct {
	UserID    UserID
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}

type AccountExport struct {
	GeneratedAt      time.Time        `json:"generated_at"`
	Account          map[string]any   `json:"account"`
	Profile          map[string]any   `json:"profile,omitempty"`
	Wishlist         []map[string]any `json:"wishlist"`
	Setups           []map[string]any `json:"setups"`
	Recommendations  []map[string]any `json:"recommendations"`
	ConsentEvents    []map[string]any `json:"consent_events"`
	AnalyticsEvents  []map[string]any `json:"analytics_events"`
	SecurityMetadata map[string]any   `json:"security_metadata"`
}
