package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	identity "rigmark/internal/modules/identity/application"
	"rigmark/internal/modules/identity/domain"
)

const maximumAuthBodyBytes = 16 * 1024

type principalContextKey struct{}

type credentialRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

type authResponse struct {
	User userResponse `json:"user"`
}

func (h *Handler) register(response http.ResponseWriter, request *http.Request) {
	input, ok := h.decodeCredentials(response, request)
	if !ok {
		return
	}
	session, err := h.auth.Register(request.Context(), input.Email, input.Password)
	if err != nil {
		h.writeAuthError(response, err)
		return
	}
	h.setSessionCookie(response, session.RawToken, session.ExpiresAt)
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusCreated, authResponse{User: userDTO(session.User)}, h.logger)
}

func (h *Handler) login(response http.ResponseWriter, request *http.Request) {
	input, ok := h.decodeCredentials(response, request)
	if !ok {
		return
	}
	session, err := h.auth.Login(request.Context(), input.Email, input.Password)
	if err != nil {
		h.writeAuthError(response, err)
		return
	}
	h.setSessionCookie(response, session.RawToken, session.ExpiresAt)
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, authResponse{User: userDTO(session.User)}, h.logger)
}

func (h *Handler) logout(response http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie(h.cookie.Name)
	if err == nil {
		if err := h.auth.Logout(request.Context(), cookie.Value); err != nil {
			h.logger.Error("log out session", "error", err)
			writeAPIError(
				response,
				http.StatusInternalServerError,
				"authentication_unavailable",
				"We could not sign you out. Please try again.",
				nil,
				h.logger,
			)
			return
		}
	}
	h.clearSessionCookie(response)
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(response http.ResponseWriter, request *http.Request) {
	principal, ok := principalFromContext(request.Context())
	if !ok {
		writeAPIError(
			response,
			http.StatusUnauthorized,
			"authentication_required",
			"Sign in to continue.",
			nil,
			h.logger,
		)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, authResponse{User: userResponse{
		ID: string(principal.UserID), Email: principal.Email, Roles: roleStrings(principal.Roles),
	}}, h.logger)
}

func (h *Handler) requireRole(role domain.Role, next http.Handler) http.Handler {
	return h.requireAnyRole([]domain.Role{role}, next)
}

func (h *Handler) requireAnyRole(roles []domain.Role, next http.Handler) http.Handler {
	return h.requireAuthentication(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		principal, ok := principalFromContext(request.Context())
		allowed := false
		if ok {
			for _, role := range roles {
				if principal.HasRole(role) {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			writeAPIError(response, http.StatusForbidden, "permission_denied", "You do not have permission to access this area.", nil, h.logger)
			return
		}
		next.ServeHTTP(response, request)
	}))
}

func roleStrings(roles []domain.Role) []string {
	values := make([]string, len(roles))
	for index, role := range roles {
		values[index] = string(role)
	}
	return values
}

func (h *Handler) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(h.cookie.Name)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			writeAPIError(
				response,
				http.StatusUnauthorized,
				"authentication_required",
				"Sign in to continue.",
				nil,
				h.logger,
			)
			return
		}
		principal, err := h.auth.Authenticate(request.Context(), cookie.Value)
		if errors.Is(err, identity.ErrUnauthenticated) {
			h.clearSessionCookie(response)
			writeAPIError(
				response,
				http.StatusUnauthorized,
				"authentication_required",
				"Sign in to continue.",
				nil,
				h.logger,
			)
			return
		}
		if err != nil {
			h.logger.Error("authenticate session", "error", err)
			writeAPIError(
				response,
				http.StatusServiceUnavailable,
				"authentication_unavailable",
				"Authentication is temporarily unavailable.",
				nil,
				h.logger,
			)
			return
		}
		next.ServeHTTP(response, request.WithContext(context.WithValue(
			request.Context(),
			principalContextKey{},
			principal,
		)))
	})
}

func (h *Handler) attachAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(h.cookie.Name)
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			next.ServeHTTP(response, request)
			return
		}
		principal, err := h.auth.Authenticate(request.Context(), cookie.Value)
		if errors.Is(err, identity.ErrUnauthenticated) {
			h.clearSessionCookie(response)
			next.ServeHTTP(response, request)
			return
		}
		if err != nil {
			h.logger.Error("optionally authenticate session", "error", err)
			writeAPIError(response, http.StatusServiceUnavailable, "authentication_unavailable",
				"Authentication is temporarily unavailable.", nil, h.logger)
			return
		}
		next.ServeHTTP(response, request.WithContext(context.WithValue(
			request.Context(), principalContextKey{}, principal,
		)))
	})
}

func principalFromContext(ctx context.Context) (domain.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(domain.Principal)
	return principal, ok
}

func (h *Handler) decodeCredentials(
	response http.ResponseWriter,
	request *http.Request,
) (credentialRequest, bool) {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(
			response,
			http.StatusUnsupportedMediaType,
			"unsupported_media_type",
			"Use application/json for this request.",
			nil,
			h.logger,
		)
		return credentialRequest{}, false
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumAuthBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var input credentialRequest
	if err := decoder.Decode(&input); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body is invalid.", nil, h.logger)
		return credentialRequest{}, false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body must contain one JSON object.", nil, h.logger)
		return credentialRequest{}, false
	}
	return input, true
}

func (h *Handler) writeAuthError(response http.ResponseWriter, err error) {
	var validationError identity.ValidationError
	switch {
	case errors.As(err, &validationError):
		writeAPIError(
			response,
			http.StatusUnprocessableEntity,
			"validation_failed",
			"Check the highlighted fields.",
			validationError.Fields,
			h.logger,
		)
	case errors.Is(err, identity.ErrEmailAlreadyUsed):
		writeAPIError(
			response,
			http.StatusConflict,
			"email_already_registered",
			"An account already exists for this email.",
			map[string]string{"email": "Sign in or use a different email."},
			h.logger,
		)
	case errors.Is(err, identity.ErrInvalidCredentials):
		writeAPIError(
			response,
			http.StatusUnauthorized,
			"invalid_credentials",
			"The email or password is incorrect.",
			nil,
			h.logger,
		)
	default:
		h.logger.Error("authentication request failed", "error", err)
		writeAPIError(
			response,
			http.StatusInternalServerError,
			"authentication_unavailable",
			"Authentication is temporarily unavailable.",
			nil,
			h.logger,
		)
	}
}

func (h *Handler) setSessionCookie(response http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(response, &http.Cookie{
		Name:     h.cookie.Name,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   h.cookie.MaxAge,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(response http.ResponseWriter) {
	http.SetCookie(response, &http.Cookie{
		Name:     h.cookie.Name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func userDTO(user domain.User) userResponse {
	return userResponse{ID: string(user.ID), Email: user.Email, Roles: roleStrings(user.Roles)}
}
