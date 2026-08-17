package httpapi

import (
	"context"
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

func TestRequireAnyRoleSeparatesEvidenceEditorAndReviewer(t *testing.T) {
	for _, test := range []struct {
		name       string
		roles      []domain.Role
		required   []domain.Role
		wantStatus int
	}{
		{name: "editor allowed on editor route", roles: []domain.Role{domain.RoleEvidenceEditor}, required: []domain.Role{domain.RoleEvidenceEditor}, wantStatus: http.StatusNoContent},
		{name: "editor denied on review route", roles: []domain.Role{domain.RoleEvidenceEditor}, required: []domain.Role{domain.RoleEvidenceReviewer}, wantStatus: http.StatusForbidden},
		{name: "reviewer allowed on review route", roles: []domain.Role{domain.RoleEvidenceReviewer}, required: []domain.Role{domain.RoleEvidenceReviewer}, wantStatus: http.StatusNoContent},
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
