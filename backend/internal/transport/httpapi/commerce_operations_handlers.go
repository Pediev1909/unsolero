package httpapi

import (
	"errors"
	"net/http"
	"strings"

	commerce "rigmark/internal/modules/commerce/application"
	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

type providerConfigurationRequest struct {
	MerchantID              string  `json:"merchant_id"`
	ProviderKey             string  `json:"provider_key"`
	AdapterKey              string  `json:"adapter_key"`
	ExternalMerchantID      string  `json:"external_merchant_id"`
	CredentialReference     *string `json:"credential_reference"`
	ScheduleIntervalMinutes int     `json:"schedule_interval_minutes"`
	FreshnessTTLMinutes     int     `json:"freshness_ttl_minutes"`
}

func (h *Handler) adminListCommerceProviders(response http.ResponseWriter, request *http.Request) {
	items, err := h.commerceOperations.ListConfigurations(request.Context())
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) adminCreateCommerceProvider(response http.ResponseWriter, request *http.Request) {
	var body providerConfigurationRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	if !validUUID(body.MerchantID) {
		h.writeCommerceOperationsError(response, commerce.ErrInvalidAttribution)
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.commerceOperations.CreateConfiguration(request.Context(), principal.UserID,
		domain.ProviderConfigurationInput{MerchantID: domain.MerchantID(body.MerchantID),
			ProviderKey: body.ProviderKey, AdapterKey: body.AdapterKey,
			ExternalMerchantID: body.ExternalMerchantID, CredentialReference: body.CredentialReference,
			ScheduleIntervalMinutes: body.ScheduleIntervalMinutes, FreshnessTTLMinutes: body.FreshnessTTLMinutes})
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, result)
}

func (h *Handler) adminSetCommerceProviderLifecycle(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("providerID"), h)
	if !ok {
		return
	}
	var body struct {
		Status domain.ProviderLifecycle `json:"status"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.commerceOperations.SetLifecycle(request.Context(), principal.UserID,
		domain.ProviderConfigurationID(id), body.Status)
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, result)
}

func (h *Handler) adminTriggerCommerceImport(response http.ResponseWriter, request *http.Request) {
	var body struct {
		ProviderConfigurationID string `json:"provider_configuration_id"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	if !validUUID(body.ProviderConfigurationID) {
		h.writeCommerceOperationsError(response, commerce.ErrInvalidAttribution)
		return
	}
	key := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
	principal, _ := principalFromContext(request.Context())
	result, err := h.commerceOperations.TriggerManual(request.Context(), principal.UserID,
		domain.ProviderConfigurationID(body.ProviderConfigurationID), key)
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusAccepted, result)
}

func (h *Handler) adminRetryCommerceImport(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("importID"), h)
	if !ok {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.commerceOperations.Retry(request.Context(), principal.UserID,
		domain.ImportRunID(id), strings.TrimSpace(request.Header.Get("Idempotency-Key")))
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusAccepted, result)
}

func (h *Handler) adminListCommerceImports(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	items, total, err := h.commerceOperations.ListImports(request.Context(), page, pageSize)
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminPage(items, total, page, pageSize))
}

func (h *Handler) adminListCommerceImportFailures(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("importID"), h)
	if !ok {
		return
	}
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	items, total, err := h.commerceOperations.ListFailures(request.Context(), domain.ImportRunID(id), page, pageSize)
	if err != nil {
		h.writeCommerceOperationsError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminPage(items, total, page, pageSize))
}

func adminPage[T any](items []T, total int64, page, pageSize int) map[string]any {
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}
	return map[string]any{"items": items, "page": page, "page_size": pageSize,
		"total": total, "total_pages": totalPages}
}

func (h *Handler) writeCommerceOperationsError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, commerce.ErrInvalidAttribution):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_commerce_operation", "The commerce operation is invalid.", nil, h.logger)
	case errors.Is(err, ports.ErrProviderDisabled):
		writeAPIError(response, http.StatusConflict, "provider_disabled", "The provider is disabled or has no verified configuration.", nil, h.logger)
	case errors.Is(err, ports.ErrProviderUnavailable):
		writeAPIError(response, http.StatusServiceUnavailable, "provider_unavailable", "The provider is temporarily unavailable.", nil, h.logger)
	case errors.Is(err, ports.ErrImportNotFound):
		writeAPIError(response, http.StatusNotFound, "commerce_record_not_found", "The commerce record was not found.", nil, h.logger)
	case errors.Is(err, ports.ErrImportConflict):
		writeAPIError(response, http.StatusConflict, "import_conflict", "The import cannot enter that state.", nil, h.logger)
	default:
		h.logger.Error("commerce operation failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "commerce_operation_failed", "The commerce operation could not be completed.", nil, h.logger)
	}
}
