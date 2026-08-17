package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	analytics "rigmark/internal/modules/analytics/application"
	analyticsdomain "rigmark/internal/modules/analytics/domain"
)

const maximumAnalyticsBodyBytes = 16 * 1024

type analyticsEventRequest struct {
	Name         string                     `json:"name"`
	Surface      string                     `json:"surface"`
	SessionID    string                     `json:"session_id"`
	ConsentState string                     `json:"consent_state"`
	Properties   map[string]json.RawMessage `json:"properties"`
	Context      analyticsContextRequest    `json:"context"`
}

type analyticsContextRequest struct {
	PagePath      *string `json:"page_path"`
	TrafficSource *string `json:"traffic_source"`
	TrafficMedium *string `json:"traffic_medium"`
	Campaign      *string `json:"campaign"`
	ReferrerHost  *string `json:"referrer_host"`
}

func (h *Handler) recordAnalyticsEvent(response http.ResponseWriter, request *http.Request) {
	var body analyticsEventRequest
	if !h.decodeAnalyticsJSON(response, request, &body) {
		return
	}
	sessionID := strings.TrimSpace(body.SessionID)
	event := analyticsdomain.Event{
		Name:          strings.TrimSpace(body.Name),
		Surface:       strings.TrimSpace(body.Surface),
		SessionID:     &sessionID,
		ConsentState:  strings.TrimSpace(body.ConsentState),
		Properties:    body.Properties,
		PagePath:      body.Context.PagePath,
		TrafficSource: body.Context.TrafficSource,
		TrafficMedium: body.Context.TrafficMedium,
		Campaign:      body.Context.Campaign,
		ReferrerHost:  body.Context.ReferrerHost,
	}
	if principal, authenticated := principalFromContext(request.Context()); authenticated {
		userID := string(principal.UserID)
		event.UserID = &userID
	} else {
		event.AnonymousID = &sessionID
	}
	if requestID := strings.TrimSpace(request.Header.Get("X-Request-ID")); requestID != "" && len(requestID) <= 128 {
		event.RequestID = &requestID
	}
	if err := h.analytics.RecordClientEvent(request.Context(), event); err != nil {
		if errors.Is(err, analytics.ErrInvalidEvent) {
			writeAPIError(response, http.StatusUnprocessableEntity, "invalid_analytics_event", "This analytics event is invalid.", nil, h.logger)
			return
		}
		h.logger.Error("record analytics event", "error", err)
		writeAPIError(response, http.StatusServiceUnavailable, "analytics_unavailable", "Analytics are temporarily unavailable.", nil, h.logger)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
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
