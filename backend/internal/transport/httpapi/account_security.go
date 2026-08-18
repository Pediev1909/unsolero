package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"

	identity "rigmark/internal/modules/identity/application"
	"rigmark/internal/platform/observability"
)

type emailRequest struct {
	Email string `json:"email"`
}
type tokenRequest struct {
	Token string `json:"token"`
}
type passwordResetCompleteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
type passwordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
type accountDeleteRequest struct {
	Password     string `json:"password"`
	Confirmation string `json:"confirmation"`
}
type passwordRequest struct {
	Password string `json:"password"`
}
type mfaCodeRequest struct {
	Code string `json:"code"`
}

func securityRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		ctx := identity.WithSecurityRequest(request.Context(), identity.SecurityRequest{
			RequestID: request.Header.Get("X-Request-ID"), Surface: "api",
		})
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func (h *Handler) requestEmailVerification(response http.ResponseWriter, request *http.Request) {
	var body emailRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if _, err := h.security.RequestEmailVerification(request.Context(), body.Email); err != nil {
		if h.metrics != nil {
			h.metrics.Increment(observability.MetricEmailDeliveryFailure)
		}
		h.logger.Error("record email verification delivery intent", "error", err)
	}
	writeJSON(response, http.StatusAccepted, map[string]any{"recorded": true,
		"message": "If the account is eligible, a verification delivery intent has been recorded."}, h.logger)
}

func (h *Handler) completeEmailVerification(response http.ResponseWriter, request *http.Request) {
	var body tokenRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.security.VerifyEmail(request.Context(), body.Token); err != nil {
		h.writeSecurityTokenError(response, err, "verification")
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"verified": true}, h.logger)
}

func (h *Handler) requestPasswordReset(response http.ResponseWriter, request *http.Request) {
	var body emailRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if _, err := h.security.RequestPasswordReset(request.Context(), body.Email); err != nil {
		if h.metrics != nil {
			h.metrics.Increment(observability.MetricEmailDeliveryFailure)
		}
		h.logger.Error("record password reset delivery intent", "error", err)
	}
	writeJSON(response, http.StatusAccepted, map[string]any{"recorded": true,
		"message": "If the account is eligible, a password-reset delivery intent has been recorded."}, h.logger)
}

func (h *Handler) completePasswordReset(response http.ResponseWriter, request *http.Request) {
	var body passwordResetCompleteRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.security.ResetPassword(request.Context(), body.Token, body.Password); err != nil {
		if validation, ok := err.(identity.ValidationError); ok {
			writeAPIError(response, http.StatusUnprocessableEntity, "validation_failed", "Check the highlighted fields.", validation.Fields, h.logger)
			return
		}
		h.writeSecurityTokenError(response, err, "password reset")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) changePassword(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	if principal.IsPrivileged() && h.securityPolicy.EnforcePrivilegedMFA && !h.security.RecentMFA(principal) {
		writeAPIError(response, http.StatusForbidden, "mfa_step_up_required", "Recent multi-factor authentication is required.", nil, h.logger)
		return
	}
	var body passwordChangeRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.security.ChangePassword(request.Context(), principal, body.CurrentPassword, body.NewPassword); err != nil {
		switch {
		case errors.Is(err, identity.ErrCurrentPassword):
			writeAPIError(response, http.StatusUnauthorized, "current_password_invalid", "The current password is incorrect.", nil, h.logger)
		case isValidationError(err):
			validation := err.(identity.ValidationError)
			writeAPIError(response, http.StatusUnprocessableEntity, "validation_failed", "Check the highlighted fields.", validation.Fields, h.logger)
		default:
			h.logger.Error("change password", "error", err)
			writeAPIError(response, http.StatusInternalServerError, "security_operation_failed", "The password could not be changed.", nil, h.logger)
		}
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listAccountSessions(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	sessions, err := h.security.ListSessions(request.Context(), principal)
	if err != nil {
		h.logger.Error("list account sessions", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "sessions_unavailable", "Active sessions could not be loaded.", nil, h.logger)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"sessions": sessions}, h.logger)
}

func (h *Handler) revokeAccountSession(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	if err := h.security.RevokeSession(request.Context(), principal, request.PathValue("sessionID")); err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			writeAPIError(response, http.StatusNotFound, "session_not_found", "The session was not found.", nil, h.logger)
		} else {
			h.logger.Error("revoke account session", "error", err)
			writeAPIError(response, http.StatusInternalServerError, "session_revoke_failed", "The session could not be revoked.", nil, h.logger)
		}
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) revokeOtherAccountSessions(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	count, err := h.security.RevokeOtherSessions(request.Context(), principal)
	if err != nil {
		h.logger.Error("revoke other sessions", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "session_revoke_failed", "Other sessions could not be revoked.", nil, h.logger)
		return
	}
	writeJSON(response, http.StatusOK, map[string]any{"revoked": count}, h.logger)
}

func (h *Handler) exportAccount(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	export, err := h.security.ExportAccount(request.Context(), principal)
	if err != nil {
		h.logger.Error("export account", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "export_unavailable", "Your account export could not be generated.", nil, h.logger)
		return
	}
	response.Header().Set("Content-Disposition", `attachment; filename="unsolero-account-export.json"`)
	writeJSON(response, http.StatusOK, export, h.logger)
}

func (h *Handler) deleteAccount(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body accountDeleteRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if principal.IsPrivileged() && h.securityPolicy.EnforcePrivilegedMFA && !h.security.RecentMFA(principal) {
		writeAPIError(response, http.StatusForbidden, "mfa_step_up_required", "Recent multi-factor authentication is required.", nil, h.logger)
		return
	}
	if err := h.security.DeleteAccount(request.Context(), principal, body.Password, body.Confirmation); err != nil {
		switch {
		case errors.Is(err, identity.ErrConfirmation):
			writeAPIError(response, http.StatusUnprocessableEntity, "confirmation_required", "Type DELETE to confirm account deletion.", nil, h.logger)
		case errors.Is(err, identity.ErrCurrentPassword):
			writeAPIError(response, http.StatusUnauthorized, "current_password_invalid", "The current password is incorrect.", nil, h.logger)
		default:
			h.logger.Error("delete account", "error", err)
			writeAPIError(response, http.StatusInternalServerError, "account_delete_failed", "The account could not be deleted.", nil, h.logger)
		}
		return
	}
	h.clearSessionCookie(response)
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) beginMFAEnrollment(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body passwordRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	value, err := h.security.BeginMFAEnrollment(request.Context(), principal, body.Password)
	if err != nil {
		h.writeMFAError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, value, h.logger)
}
func (h *Handler) confirmMFAEnrollment(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body mfaCodeRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	value, err := h.security.ConfirmMFAEnrollment(request.Context(), principal, body.Code)
	if err != nil {
		h.writeMFAError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value, h.logger)
}
func (h *Handler) regenerateRecoveryCodes(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body mfaCodeRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	value, err := h.security.RegenerateRecoveryCodes(request.Context(), principal, body.Code)
	if err != nil {
		h.writeMFAError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, value, h.logger)
}
func (h *Handler) stepUpMFA(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body mfaCodeRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.security.VerifyStepUp(request.Context(), principal, body.Code); err != nil {
		h.writeMFAError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) completeMFALogin(response http.ResponseWriter, request *http.Request) {
	var body mfaCodeRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	challenge, err := request.Cookie(h.mfaCookieName())
	if err != nil || strings.TrimSpace(challenge.Value) == "" {
		writeAPIError(response, http.StatusUnauthorized, "mfa_challenge_invalid", "The sign-in challenge is invalid or expired.", nil, h.logger)
		return
	}
	session, err := h.security.CompleteLoginMFA(request.Context(), challenge.Value, body.Code)
	if err != nil {
		h.writeMFAError(response, err)
		return
	}
	h.clearMFACookie(response)
	h.setSessionCookie(response, session.RawToken, session.ExpiresAt)
	user := userDTO(session.User)
	user.MFAEnabled = true
	writeJSON(response, http.StatusOK, authResponse{User: user}, h.logger)
}

func (h *Handler) listDevelopmentEmails(response http.ResponseWriter, request *http.Request) {
	recipient := strings.TrimSpace(strings.ToLower(request.URL.Query().Get("recipient")))
	writeJSON(response, http.StatusOK, map[string]any{"delivery": "development_intent_only", "messages": h.developmentEmail.Messages(recipient)}, h.logger)
}

func (h *Handler) writeSecurityTokenError(response http.ResponseWriter, err error, operation string) {
	switch {
	case errors.Is(err, identity.ErrExpiredToken):
		writeAPIError(response, http.StatusGone, "token_expired", "This "+operation+" link has expired.", nil, h.logger)
	case errors.Is(err, identity.ErrUsedToken):
		writeAPIError(response, http.StatusGone, "token_used", "This "+operation+" link has already been used.", nil, h.logger)
	case errors.Is(err, identity.ErrInvalidToken):
		writeAPIError(response, http.StatusBadRequest, "token_invalid", "This "+operation+" link is invalid.", nil, h.logger)
	default:
		h.logger.Error("security token operation", "operation", operation, "error", err)
		writeAPIError(response, http.StatusInternalServerError, "security_operation_failed", "The security operation could not be completed.", nil, h.logger)
	}
}
func (h *Handler) writeMFAError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrCurrentPassword):
		writeAPIError(response, http.StatusUnauthorized, "current_password_invalid", "The current password is incorrect.", nil, h.logger)
	case errors.Is(err, identity.ErrInvalidMFACode):
		writeAPIError(response, http.StatusUnauthorized, "mfa_code_invalid", "The authentication code is invalid.", nil, h.logger)
	case errors.Is(err, identity.ErrMFAChallenge):
		writeAPIError(response, http.StatusUnauthorized, "mfa_challenge_invalid", "The sign-in challenge is invalid or expired.", nil, h.logger)
	case errors.Is(err, identity.ErrMFAUnavailable):
		writeAPIError(response, http.StatusConflict, "mfa_not_configured", "Multi-factor authentication is not configured.", nil, h.logger)
	default:
		h.logger.Error("MFA operation", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "mfa_unavailable", "Multi-factor authentication is temporarily unavailable.", nil, h.logger)
	}
}
func isValidationError(err error) bool {
	var value identity.ValidationError
	return errors.As(err, &value)
}

func decodeStrictJSON(response http.ResponseWriter, request *http.Request, destination any, maximumBytes int64, logger *slog.Logger) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(response, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use application/json for this request.", nil, logger)
		return false
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body is invalid.", nil, logger)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body must contain one JSON object.", nil, logger)
		return false
	}
	return true
}
