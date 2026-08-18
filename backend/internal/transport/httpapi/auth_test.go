package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identity "rigmark/internal/modules/identity/application"
	"rigmark/internal/modules/identity/domain"
)

var testCookieConfig = AuthCookieConfig{Name: "test_session", Secure: true, MaxAge: 3600}

type authStub struct {
	session            identity.AuthenticatedSession
	principal          domain.Principal
	err                error
	authenticatedToken string
	loggedOutToken     string
}

type accountSecurityStub struct {
	AccountSecurityService
	verificationRequests int
	recentMFA            bool
}

func (stub *accountSecurityStub) RequestEmailVerification(context.Context, string) (identity.RequestReceipt, error) {
	stub.verificationRequests++
	return identity.RequestReceipt{Recorded: true}, nil
}

func (stub *accountSecurityStub) RecentMFA(domain.Principal) bool { return stub.recentMFA }
func (stub *accountSecurityStub) RecordAuthorizationFailure(context.Context, domain.Principal, domain.Permission) error {
	return nil
}

func (stub authStub) Register(
	context.Context,
	string,
	string,
) (identity.AuthenticatedSession, error) {
	return stub.session, stub.err
}

func (stub authStub) Login(
	context.Context,
	string,
	string,
) (identity.AuthenticatedSession, error) {
	return stub.session, stub.err
}

func (stub *authStub) Logout(_ context.Context, token string) error {
	stub.loggedOutToken = token
	return stub.err
}

func (stub *authStub) Authenticate(
	_ context.Context,
	token string,
) (domain.Principal, error) {
	stub.authenticatedToken = token
	return stub.principal, stub.err
}

func newAuthTestRouter(authService AuthenticationService) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(healthStub{}, authService, testCookieConfig, logger)
}

func TestRegisterSetsProtectedSessionCookieWithoutExposingSecrets(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	authService := &authStub{session: identity.AuthenticatedSession{
		User:      domain.User{ID: "user-1", Email: "person@example.com"},
		RawToken:  "opaque-secret-token",
		ExpiresAt: expiresAt,
	}}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(`{"email":"person@example.com","password":"a long secure password"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newAuthTestRouter(authService).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookie count = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != testCookieConfig.Name || !cookie.HttpOnly || !cookie.Secure ||
		cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" {
		t.Fatalf("session cookie = %#v", cookie)
	}
	body := response.Body.String()
	if strings.Contains(body, "opaque-secret-token") || strings.Contains(body, "password") {
		t.Fatalf("response exposed authentication secret: %s", body)
	}
	if !strings.Contains(body, `"email":"person@example.com"`) {
		t.Fatalf("response body = %s", body)
	}
}

func TestLoginReturnsSafeCredentialError(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/login",
		strings.NewReader(`{"email":"person@example.com","password":"wrong password"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newAuthTestRouter(&authStub{err: identity.ErrInvalidCredentials}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) {
		t.Fatalf("body = %s", response.Body)
	}
}

func TestRegistrationIsAntiEnumeratedWhenSecurityServiceIsEnabled(t *testing.T) {
	security := &accountSecurityStub{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	responses := make([]*httptest.ResponseRecorder, 0, 2)
	for _, authService := range []*authStub{
		{session: identity.AuthenticatedSession{User: domain.User{ID: "new-user"}, RawToken: "new-token", ExpiresAt: time.Now().Add(time.Hour)}},
		{err: identity.ErrEmailAlreadyUsed},
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{"email":"person@example.com","password":"a long secure password"}`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		NewRouter(healthStub{}, authService, testCookieConfig, logger, PublicServices{Security: security}).ServeHTTP(response, request)
		responses = append(responses, response)
	}
	if responses[0].Code != http.StatusAccepted || responses[1].Code != http.StatusAccepted || responses[0].Body.String() != responses[1].Body.String() {
		t.Fatalf("registration responses differ: %d %q vs %d %q", responses[0].Code, responses[0].Body, responses[1].Code, responses[1].Body)
	}
	if len(responses[0].Result().Cookies()) != 0 || len(responses[1].Result().Cookies()) != 0 {
		t.Fatal("anti-enumerated registration must not create a browser session")
	}
}

func TestLoginMFAChallengeUsesScopedHttpOnlyCookie(t *testing.T) {
	expires := time.Now().Add(5 * time.Minute)
	authService := &authStub{session: identity.AuthenticatedSession{MFARequired: true,
		MFAChallengeToken: "challenge-secret", MFAChallengeExpiresAt: &expires}}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"a long secure password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newAuthTestRouter(authService).ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "test_session_mfa" || cookies[0].Path != "/api/auth/mfa/complete" ||
		!cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteStrictMode {
		t.Fatalf("MFA challenge cookie = %#v", cookies)
	}
	if strings.Contains(response.Body.String(), "challenge-secret") {
		t.Fatal("MFA challenge token leaked into response")
	}
}

func TestMeRequiresAndResolvesAuthenticatedSession(t *testing.T) {
	authService := &authStub{principal: domain.Principal{
		UserID: "user-1", Email: "person@example.com",
	}}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "opaque-secret-token"})
	response := httptest.NewRecorder()

	newAuthTestRouter(authService).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}
	if authService.authenticatedToken != "opaque-secret-token" {
		t.Errorf("authenticated token = %q", authService.authenticatedToken)
	}
	if strings.Contains(response.Body.String(), "opaque-secret-token") {
		t.Fatal("me response exposed the session token")
	}
}

func TestMeRejectsMissingSession(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	response := httptest.NewRecorder()

	newAuthTestRouter(&authStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestLogoutRevokesAndClearsSession(t *testing.T) {
	authService := &authStub{}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "opaque-secret-token"})
	response := httptest.NewRecorder()

	newAuthTestRouter(authService).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if authService.loggedOutToken != "opaque-secret-token" {
		t.Errorf("logged out token = %q", authService.loggedOutToken)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Fatalf("cleared cookies = %#v", cookies)
	}
}

func TestAuthRejectsCrossOriginMutation(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"http://rigmark.test/api/auth/login",
		strings.NewReader(`{"email":"person@example.com","password":"a long secure password"}`),
	)
	request.Header.Set("Origin", "https://attacker.example")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newAuthTestRouter(&authStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if !strings.Contains(response.Body.String(), `"code":"origin_not_allowed"`) {
		t.Fatalf("body = %s", response.Body)
	}
}

func TestAuthRejectsUnknownJSONFields(t *testing.T) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/register",
		strings.NewReader(`{"email":"person@example.com","password":"a long secure password","admin":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newAuthTestRouter(&authStub{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestRequireAdminRoleRejectsAuthenticatedMember(t *testing.T) {
	authService := &authStub{principal: domain.Principal{UserID: "user-1", Email: "member@example.com"}}
	handler := &Handler{
		auth:   authService,
		cookie: testCookieConfig,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	protected := handler.requireRole(domain.RoleAdmin, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "member-token"})
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestRequireAdminRoleAllowsAdministrator(t *testing.T) {
	authService := &authStub{principal: domain.Principal{UserID: "user-1", Email: "admin@example.com", Roles: []domain.Role{domain.RoleAdmin}}}
	handler := &Handler{
		auth:   authService,
		cookie: testCookieConfig,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	protected := handler.requireRole(domain.RoleAdmin, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "admin-token"})
	response := httptest.NewRecorder()

	protected.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestAffiliateOptionalAuthenticationFailsOpen(t *testing.T) {
	authService := &authStub{err: errors.New("session store unavailable")}
	handler := &Handler{auth: authService, cookie: testCookieConfig,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	called := false
	protected := handler.attachOptionalAuthenticationFailOpen(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		called = true
		if _, authenticated := principalFromContext(request.Context()); authenticated {
			t.Fatal("failed authentication must continue anonymously")
		}
		response.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodGet, "/api/affiliate/click/offer", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "opaque-token"})
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	if !called || response.Code != http.StatusNoContent {
		t.Fatalf("called=%v status=%d", called, response.Code)
	}
}

func TestRequireAnyRoleSeparatesGovernanceEditorsAndReviewers(t *testing.T) {
	for _, test := range []struct {
		name       string
		roles      []domain.Role
		required   []domain.Role
		wantStatus int
	}{
		{name: "editor allowed on editor route", roles: []domain.Role{domain.RoleEvidenceEditor}, required: []domain.Role{domain.RoleEvidenceEditor}, wantStatus: http.StatusNoContent},
		{name: "editor denied on review route", roles: []domain.Role{domain.RoleEvidenceEditor}, required: []domain.Role{domain.RoleEvidenceReviewer}, wantStatus: http.StatusForbidden},
		{name: "reviewer allowed on review route", roles: []domain.Role{domain.RoleEvidenceReviewer}, required: []domain.Role{domain.RoleEvidenceReviewer}, wantStatus: http.StatusNoContent},
		{name: "policy editor allowed on editor route", roles: []domain.Role{domain.RolePolicyEditor}, required: []domain.Role{domain.RolePolicyEditor}, wantStatus: http.StatusNoContent},
		{name: "policy editor denied on review route", roles: []domain.Role{domain.RolePolicyEditor}, required: []domain.Role{domain.RolePolicyReviewer}, wantStatus: http.StatusForbidden},
		{name: "policy reviewer denied on editor route", roles: []domain.Role{domain.RolePolicyReviewer}, required: []domain.Role{domain.RolePolicyEditor}, wantStatus: http.StatusForbidden},
		{name: "policy reviewer allowed on review route", roles: []domain.Role{domain.RolePolicyReviewer}, required: []domain.Role{domain.RolePolicyReviewer}, wantStatus: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			authService := &authStub{principal: domain.Principal{UserID: "user-1", Email: "evidence@example.invalid", Roles: test.roles}}
			handler := &Handler{auth: authService, cookie: testCookieConfig, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			protected := handler.requireAnyRole(test.required, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodPost, "/api/admin/evidence", nil)
			request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "token"})
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}

func TestRequirePermissionEnforcesLeastPrivilege(t *testing.T) {
	tests := []struct {
		name       string
		role       domain.Role
		permission domain.Permission
		want       int
	}{
		{"catalog edits catalog", domain.RoleCatalogEditor, domain.PermissionCatalogUpdate, http.StatusNoContent},
		{"catalog cannot approve evidence", domain.RoleCatalogEditor, domain.PermissionEvidenceApprove, http.StatusForbidden},
		{"reviewer approves evidence", domain.RoleEvidenceReviewer, domain.PermissionEvidenceApprove, http.StatusNoContent},
		{"commerce cannot mutate policy", domain.RoleCommerceOperator, domain.PermissionPolicyCreate, http.StatusForbidden},
		{"analyst cannot mutate commerce", domain.RoleAnalyst, domain.PermissionCommerceUpdate, http.StatusForbidden},
		{"analyst reads aggregates", domain.RoleAnalyst, domain.PermissionAnalyticsRead, http.StatusNoContent},
		{"analyst cannot read raw events", domain.RoleAnalyst, domain.PermissionAnalyticsRawRead, http.StatusForbidden},
		{"analyst cannot read consent records", domain.RoleAnalyst, domain.PermissionConsentRead, http.StatusForbidden},
		{"analyst cannot read security events", domain.RoleAnalyst, domain.PermissionSecurityEventsRead, http.StatusForbidden},
		{"content cannot export analytics", domain.RoleContentEditor, domain.PermissionAnalyticsExport, http.StatusForbidden},
		{"unknown forged role denied", domain.Role("administrator"), domain.PermissionAdminRead, http.StatusForbidden},
		{"administrator allowed", domain.RoleAdmin, domain.PermissionSecurityEventsRead, http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authService := &authStub{principal: domain.Principal{UserID: "user-1", Roles: []domain.Role{test.role}}}
			handler := &Handler{auth: authService, cookie: testCookieConfig, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
			protected := handler.requirePermission(test.permission, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) }))
			request := httptest.NewRequest(http.MethodPost, "/api/admin/resource", nil)
			request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "session"})
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, test.want, response.Body)
			}
		})
	}
}

func TestPrivilegedPermissionRequiresBackendVerifiedRecentMFA(t *testing.T) {
	security := &accountSecurityStub{recentMFA: false}
	verifiedAt := time.Now().Add(-time.Hour)
	authService := &authStub{principal: domain.Principal{UserID: "admin", Roles: []domain.Role{domain.RoleAdmin}, EmailVerifiedAt: &verifiedAt}}
	handler := &Handler{auth: authService, security: security, securityPolicy: SecurityPolicyConfig{EnforcePrivilegedMFA: true},
		cookie: testCookieConfig, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	protected := handler.requirePermission(domain.PermissionCommerceRead, http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) }))
	request := httptest.NewRequest(http.MethodGet, "/api/admin/commerce/conversions", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "session"})
	response := httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "mfa_step_up_required") {
		t.Fatalf("stale MFA response=%d body=%s", response.Code, response.Body)
	}
	security.recentMFA = true
	response = httptest.NewRecorder()
	protected.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("recent MFA status=%d body=%s", response.Code, response.Body)
	}
}
