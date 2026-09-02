package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	newsletter "rigmark/internal/modules/newsletter/application"
	"rigmark/internal/platform/observability"
)

type NewsletterService interface {
	Subscribe(context.Context, string, string) (newsletter.Receipt, error)
	Confirm(context.Context, string) error
	Unsubscribe(context.Context, string) error
}

type newsletterSubscriptionRequest struct {
	Email  string `json:"email"`
	Source string `json:"source"`
}

// newsletterHandler owns the three public newsletter routes. It is a separate
// small type rather than more methods on Handler so the module can be wired
// with one field on PublicServices and three mux lines.
type newsletterHandler struct {
	service NewsletterService
	logger  *slog.Logger
	metrics observability.Recorder
}

func newNewsletterHandler(service NewsletterService, logger *slog.Logger, metrics observability.Recorder) *newsletterHandler {
	if metrics == nil {
		metrics = observability.DisabledRecorder{}
	}
	return &newsletterHandler{service: service, logger: logger, metrics: metrics}
}

// subscribe answers 202 with the same body for a new, refreshed, or already
// confirmed address. Only malformed input (400) and a failure to write the row
// (500) look different, and neither reveals whether the address is known.
func (h *newsletterHandler) subscribe(response http.ResponseWriter, request *http.Request) {
	var body newsletterSubscriptionRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	receipt, err := h.service.Subscribe(request.Context(), body.Email, body.Source)
	if err != nil {
		var validation newsletter.ValidationError
		if errors.As(err, &validation) {
			writeAPIError(response, http.StatusBadRequest, "validation_failed", "Check the highlighted fields.", validation.Fields, h.logger)
			return
		}
		if !receipt.Recorded {
			h.logger.Error("record newsletter subscription", "error", err)
			writeAPIError(response, http.StatusInternalServerError, "newsletter_unavailable", "The subscription could not be recorded. Please try again.", nil, h.logger)
			return
		}
		// The row exists and the person will simply not receive the email. That
		// is the operator's problem to notice, not a signal to the requester.
		h.metrics.Increment(observability.MetricEmailDeliveryFailure)
		h.logger.Error("deliver newsletter confirmation", "error", err)
	}
	writeJSON(response, http.StatusAccepted, map[string]any{"recorded": true}, h.logger)
}

func (h *newsletterHandler) confirm(response http.ResponseWriter, request *http.Request) {
	var body tokenRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.service.Confirm(request.Context(), body.Token); err != nil {
		h.writeTokenError(response, err, "confirm newsletter subscription")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *newsletterHandler) unsubscribe(response http.ResponseWriter, request *http.Request) {
	var body tokenRequest
	if !decodeStrictJSON(response, request, &body, maximumAuthBodyBytes, h.logger) {
		return
	}
	if err := h.service.Unsubscribe(request.Context(), body.Token); err != nil {
		h.writeTokenError(response, err, "unsubscribe newsletter address")
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *newsletterHandler) writeTokenError(response http.ResponseWriter, err error, operation string) {
	if errors.Is(err, newsletter.ErrInvalidToken) {
		writeAPIError(response, http.StatusBadRequest, "invalid_token", "This link is invalid or has expired.", nil, h.logger)
		return
	}
	h.logger.Error(operation, "error", err)
	writeAPIError(response, http.StatusInternalServerError, "newsletter_unavailable", "The request could not be completed. Please try again.", nil, h.logger)
}
