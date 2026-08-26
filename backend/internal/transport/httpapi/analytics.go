package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	analytics "rigmark/internal/modules/analytics/application"
	analyticsdomain "rigmark/internal/modules/analytics/domain"
	analyticsports "rigmark/internal/modules/analytics/ports"
	commercedomain "rigmark/internal/modules/commerce/domain"
)

const maximumAnalyticsBodyBytes = 16 * 1024
const defaultAnalyticsSubjectCookie = "unsolero_analytics_subject"

type analyticsEventRequest struct {
	EventID        string                     `json:"event_id"`
	Name           string                     `json:"name"`
	Surface        string                     `json:"surface"`
	SessionID      string                     `json:"session_id"`
	ConsentVersion string                     `json:"consent_version"`
	Properties     map[string]json.RawMessage `json:"properties"`
	Context        analyticsContextRequest    `json:"context"`
}

type analyticsContextRequest struct {
	PagePath      *string `json:"page_path"`
	TrafficSource *string `json:"traffic_source"`
	TrafficMedium *string `json:"traffic_medium"`
	Campaign      *string `json:"campaign"`
	ReferrerHost  *string `json:"referrer_host"`
}

type analyticsConsentRequest struct {
	State         string `json:"state"`
	PolicyVersion string `json:"policy_version"`
	Source        string `json:"source"`
}

type analyticsConsentResponse struct {
	State         string     `json:"state"`
	PolicyVersion string     `json:"policy_version,omitempty"`
	Source        string     `json:"source,omitempty"`
	DecidedAt     *time.Time `json:"decided_at,omitempty"`
}

func (h *Handler) recordAnalyticsEvent(response http.ResponseWriter, request *http.Request) {
	var body analyticsEventRequest
	if !h.decodeAnalyticsJSON(response, request, &body) {
		return
	}
	sessionID := strings.TrimSpace(body.SessionID)
	event := analyticsdomain.Event{
		ID: analyticsdomain.EventID(strings.TrimSpace(body.EventID)), Name: strings.TrimSpace(body.Name),
		Surface: strings.TrimSpace(body.Surface), SessionID: &sessionID,
		ConsentPolicyVersion: strings.TrimSpace(body.ConsentVersion), Properties: body.Properties,
		PagePath: body.Context.PagePath, TrafficSource: body.Context.TrafficSource,
		TrafficMedium: body.Context.TrafficMedium, Campaign: body.Context.Campaign,
		ReferrerHost: body.Context.ReferrerHost, Classification: analyticsTrafficClassification(request),
	}
	if principal, authenticated := principalFromContext(request.Context()); authenticated {
		userID := string(principal.UserID)
		event.UserID = &userID
	} else if subjectHash, ok := h.analyticsSubject(request); ok {
		event.AnonymousSubjectHash = subjectHash
	}
	if requestID := strings.TrimSpace(request.Header.Get("X-Request-ID")); requestID != "" && len(requestID) <= 128 {
		event.RequestID = &requestID
	}
	_, err := h.analytics.RecordClientEvent(request.Context(), event)
	if err != nil {
		switch {
		case errors.Is(err, analytics.ErrInvalidEvent):
			writeAPIError(response, http.StatusUnprocessableEntity, "invalid_analytics_event", "This analytics event is invalid.", nil, h.logger)
		case errors.Is(err, analytics.ErrConsentRequired):
			writeAPIError(response, http.StatusForbidden, "analytics_consent_required", "Analytics consent is not active for this browser or account.", nil, h.logger)
		default:
			h.logger.Error("record analytics event", "error", err)
			writeAPIError(response, http.StatusServiceUnavailable, "analytics_unavailable", "Analytics are temporarily unavailable.", nil, h.logger)
		}
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getAnalyticsConsent(response http.ResponseWriter, request *http.Request) {
	subject, ok := h.analyticsSubjectFromRequest(request)
	if !ok {
		h.writeAnalyticsConsent(response, analyticsConsentResponse{State: "unknown"})
		return
	}
	consent, err := h.analytics.GetConsent(request.Context(), subject)
	if errors.Is(err, analyticsports.ErrConsentNotFound) {
		h.writeAnalyticsConsent(response, analyticsConsentResponse{State: "unknown"})
		return
	}
	if err != nil {
		h.logger.Error("read analytics consent", "error", err)
		writeAPIError(response, http.StatusServiceUnavailable, "analytics_consent_unavailable", "Analytics preferences are temporarily unavailable.", nil, h.logger)
		return
	}
	h.writeAnalyticsConsent(response, consentDTO(consent))
}

func (h *Handler) setAnalyticsConsent(response http.ResponseWriter, request *http.Request) {
	var body analyticsConsentRequest
	if !h.decodeAnalyticsJSON(response, request, &body) {
		return
	}
	subject, ok := h.analyticsSubjectFromRequest(request)
	if !ok {
		var err error
		subject.AnonymousSubjectHash, err = h.issueAnalyticsSubject(response)
		if err != nil {
			h.logger.Error("create analytics subject", "error", err)
			writeAPIError(response, http.StatusServiceUnavailable, "analytics_consent_unavailable", "Analytics preferences are temporarily unavailable.", nil, h.logger)
			return
		}
	}
	consent, err := h.analytics.SetConsent(request.Context(), analyticsdomain.ConsentDecision{
		Subject: subject, RequestedState: strings.TrimSpace(body.State),
		PolicyVersion: strings.TrimSpace(body.PolicyVersion), Source: strings.TrimSpace(body.Source),
	})
	if errors.Is(err, analytics.ErrInvalidConsent) {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_analytics_consent", "This analytics preference is invalid.", nil, h.logger)
		return
	}
	if err != nil {
		h.logger.Error("persist analytics consent", "error", err)
		writeAPIError(response, http.StatusServiceUnavailable, "analytics_consent_unavailable", "Analytics preferences are temporarily unavailable.", nil, h.logger)
		return
	}
	h.writeAnalyticsConsent(response, consentDTO(consent))
}

func (h *Handler) claimAnalyticsIdentity(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	subjectHash, ok := h.analyticsSubject(request)
	if !ok {
		writeAPIError(response, http.StatusConflict, "analytics_identity_unavailable", "There is no eligible anonymous analytics identity to associate.", nil, h.logger)
		return
	}
	err := h.analytics.ClaimIdentity(request.Context(), subjectHash, string(principal.UserID))
	if errors.Is(err, analyticsports.ErrIdentityClaimConflict) || errors.Is(err, analyticsports.ErrIdentityClaimNotAllowed) {
		// Deliberately generic: callers cannot use this endpoint to enumerate
		// event IDs, browser subjects, or prior account associations.
		writeAPIError(response, http.StatusConflict, "analytics_identity_unavailable", "This analytics identity cannot be associated.", nil, h.logger)
		return
	}
	if err != nil {
		h.logger.Error("claim analytics identity", "error", err)
		writeAPIError(response, http.StatusServiceUnavailable, "analytics_identity_unavailable", "This analytics identity cannot be associated right now.", nil, h.logger)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) analyticsSubjectFromRequest(request *http.Request) (analyticsdomain.Subject, bool) {
	if principal, authenticated := principalFromContext(request.Context()); authenticated {
		userID := string(principal.UserID)
		return analyticsdomain.Subject{UserID: &userID}, true
	}
	hash, ok := h.analyticsSubject(request)
	return analyticsdomain.Subject{AnonymousSubjectHash: hash}, ok
}

func (h *Handler) analyticsSubject(request *http.Request) ([]byte, bool) {
	cookie, err := request.Cookie(h.analyticsSubjectCookieName())
	if err != nil {
		return nil, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil || len(raw) != 32 {
		return nil, false
	}
	hash := sha256.Sum256(raw)
	return hash[:], true
}

func (h *Handler) issueAnalyticsSubject(response http.ResponseWriter) ([]byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	maxAge := h.cookie.AnalyticsSubjectMaxAge
	if maxAge <= 0 {
		maxAge = 397 * 24 * 60 * 60
	}
	// nosemgrep: go.lang.security.audit.net.cookie-missing-secure.cookie-missing-secure -- production/staging configuration rejects Secure=false; local HTTP development remains supported.
	http.SetCookie(response, &http.Cookie{
		Name: h.analyticsSubjectCookieName(), Value: base64.RawURLEncoding.EncodeToString(raw),
		Path: "/", HttpOnly: true, Secure: h.cookie.Secure, SameSite: http.SameSiteLaxMode,
		MaxAge: maxAge,
	})
	hash := sha256.Sum256(raw)
	return hash[:], nil
}

func (h *Handler) analyticsSubjectCookieName() string {
	if h.cookie.AnalyticsSubjectName != "" {
		return h.cookie.AnalyticsSubjectName
	}
	return defaultAnalyticsSubjectCookie
}

func analyticsTrafficClassification(request *http.Request) string {
	classification := commercedomain.ClassifyClick(request.UserAgent(), request.Header.Get("Purpose"), request.Header.Get("Sec-Purpose"), request.Header.Get("X-Moz"))
	return string(classification)
}

func consentDTO(consent analyticsdomain.Consent) analyticsConsentResponse {
	decidedAt := consent.DecidedAt
	return analyticsConsentResponse{State: consent.State, PolicyVersion: consent.PolicyVersion, Source: consent.Source, DecidedAt: &decidedAt}
}

func (h *Handler) writeAnalyticsConsent(response http.ResponseWriter, value analyticsConsentResponse) {
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, value, h.logger)
}

func (h *Handler) decodeAnalyticsJSON(response http.ResponseWriter, request *http.Request, destination any) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(response, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use application/json for this request.", nil, h.logger)
		return false
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumAnalyticsBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(destination); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body is invalid.", nil, h.logger)
		return false
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body must contain one JSON object.", nil, h.logger)
		return false
	}
	return true
}
