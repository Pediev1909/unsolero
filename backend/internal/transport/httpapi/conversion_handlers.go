package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	commerce "rigmark/internal/modules/commerce/application"
	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

const maximumConversionWebhookBytes = 256 << 10

func (h *Handler) commerceConversionWebhook(response http.ResponseWriter, request *http.Request) {
	configurationID := request.PathValue("providerConfigurationID")
	if !validUUID(configurationID) {
		h.writeConversionError(response, ports.ErrWebhookRejected)
		return
	}
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(response, http.StatusUnsupportedMediaType, "unsupported_media_type", "The webhook content type is not supported.", nil, h.logger)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumConversionWebhookBytes)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		writeAPIError(response, http.StatusRequestEntityTooLarge, "webhook_too_large", "The webhook body exceeds the accepted limit.", nil, h.logger)
		return
	}
	result, err := h.conversionOperations.IngestWebhook(request.Context(),
		domain.ProviderConfigurationID(configurationID), domain.WebhookRequest{
			Headers: request.Header.Clone(), Body: body, ReceivedAt: time.Now().UTC(),
		})
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	status := http.StatusAccepted
	if result.Duplicate {
		status = http.StatusOK
	}
	h.writeAdminJSON(response, status, result)
}

func (h *Handler) adminSetConversionProvider(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("providerID"), h)
	if !ok {
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.conversionOperations.SetProviderEnabled(request.Context(), principal.UserID,
		domain.ProviderConfigurationID(id), body.Enabled)
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, result)
}

func (h *Handler) adminTriggerConversionImport(response http.ResponseWriter, request *http.Request) {
	var body struct {
		ProviderConfigurationID string `json:"provider_configuration_id"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	if !validUUID(body.ProviderConfigurationID) {
		h.writeConversionError(response, commerce.ErrInvalidAttribution)
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.conversionOperations.TriggerManualImport(request.Context(), principal.UserID,
		domain.ProviderConfigurationID(body.ProviderConfigurationID), strings.TrimSpace(request.Header.Get("Idempotency-Key")))
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusAccepted, result)
}

func (h *Handler) adminRetryConversionImport(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("importID"), h)
	if !ok {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.conversionOperations.RetryImport(request.Context(), principal.UserID,
		domain.ConversionImportRunID(id), strings.TrimSpace(request.Header.Get("Idempotency-Key")))
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusAccepted, result)
}

func (h *Handler) adminListConversions(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	query := request.URL.Query()
	filter := domain.ConversionFilter{Provider: query.Get("provider"),
		OrderStatus:          domain.OrderStatus(query.Get("order_status")),
		CommissionStatus:     domain.CommissionStatus(query.Get("commission_status")),
		AttributionStatus:    query.Get("attribution_status"),
		ReconciliationStatus: query.Get("reconciliation_status"), Currency: query.Get("currency"),
		Limit: pageSize, Offset: (page - 1) * pageSize}
	items, total, err := h.conversionOperations.ListConversions(request.Context(), filter)
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminPage(items, total, page, pageSize))
}

func (h *Handler) adminListConversionImports(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	items, total, err := h.conversionOperations.ListImports(request.Context(), page, pageSize)
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminPage(items, total, page, pageSize))
}

func (h *Handler) adminListConversionReconciliations(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	items, total, err := h.conversionOperations.ListReconciliations(request.Context(), page, pageSize)
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminPage(items, total, page, pageSize))
}

func (h *Handler) adminReconcileConversionImport(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("importID"), h)
	if !ok {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.conversionOperations.Reconcile(request.Context(), principal.UserID,
		domain.ConversionImportRunID(id), strings.TrimSpace(request.Header.Get("Idempotency-Key")))
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, result)
}

func (h *Handler) adminMonetizationMetrics(response http.ResponseWriter, request *http.Request) {
	end := time.Now().UTC()
	start := end.Add(-30 * 24 * time.Hour)
	var err error
	if raw := request.URL.Query().Get("start"); raw != "" {
		start, err = time.Parse(time.RFC3339, raw)
	}
	if err == nil {
		if raw := request.URL.Query().Get("end"); raw != "" {
			end, err = time.Parse(time.RFC3339, raw)
		}
	}
	if err != nil {
		h.writeConversionError(response, commerce.ErrInvalidAttribution)
		return
	}
	result, err := h.conversionOperations.Metrics(request.Context(), start, end)
	if err != nil {
		h.writeConversionError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, result)
}

func (h *Handler) writeConversionError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ports.ErrWebhookRejected), errors.Is(err, ports.ErrProviderDisabled):
		writeAPIError(response, http.StatusUnauthorized, "conversion_verification_failed", "The conversion event could not be verified.", nil, h.logger)
	case errors.Is(err, ports.ErrWebhookReplay):
		writeAPIError(response, http.StatusConflict, "webhook_replayed", "The webhook was already received.", nil, h.logger)
	case errors.Is(err, ports.ErrConversionConflict), errors.Is(err, ports.ErrImportConflict):
		writeAPIError(response, http.StatusConflict, "conversion_conflict", "The conversion conflicts with verified history.", nil, h.logger)
	case errors.Is(err, ports.ErrConversionNotFound), errors.Is(err, ports.ErrImportNotFound):
		writeAPIError(response, http.StatusNotFound, "conversion_not_found", "The conversion record was not found.", nil, h.logger)
	case errors.Is(err, commerce.ErrInvalidAttribution):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_conversion_operation", "The conversion operation is invalid.", nil, h.logger)
	default:
		h.logger.Error("conversion operation failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "conversion_operation_failed", "The conversion operation could not be completed.", nil, h.logger)
	}
}
