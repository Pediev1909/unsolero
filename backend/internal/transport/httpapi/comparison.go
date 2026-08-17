package httpapi

import (
	"context"
	"errors"
	"net/http"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/application"
	planningports "rigmark/internal/modules/planning/ports"
)

type ComparisonService interface {
	List(context.Context, identity.UserID) ([]catalog.ProductID, error)
	Replace(context.Context, identity.UserID, []catalog.ProductID) error
}

type comparisonRequest struct {
	ProductIDs []string `json:"product_ids"`
}

type comparisonResponse struct {
	ProductIDs []string `json:"product_ids"`
}

func (h *Handler) listComparison(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	productIDs, err := h.comparison.List(request.Context(), principal.UserID)
	if err != nil {
		h.writeComparisonError(response, err)
		return
	}
	result := make([]string, 0, len(productIDs))
	for _, productID := range productIDs {
		result = append(result, string(productID))
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, comparisonResponse{ProductIDs: result}, h.logger)
}

func (h *Handler) replaceComparison(response http.ResponseWriter, request *http.Request) {
	principal, _ := principalFromContext(request.Context())
	var body comparisonRequest
	if !h.decodeRecommendationJSON(response, request, &body) {
		return
	}
	productIDs := make([]catalog.ProductID, 0, len(body.ProductIDs))
	for _, productID := range body.ProductIDs {
		if !validUUID(productID) {
			h.writeComparisonError(response, planning.ErrInvalidComparison)
			return
		}
		productIDs = append(productIDs, catalog.ProductID(productID))
	}
	if err := h.comparison.Replace(request.Context(), principal.UserID, productIDs); err != nil {
		h.writeComparisonError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, comparisonResponse{ProductIDs: body.ProductIDs}, h.logger)
}

func (h *Handler) writeComparisonError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, planning.ErrInvalidComparison):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_comparison", "Choose up to four different products.", nil, h.logger)
	case errors.Is(err, planningports.ErrProductNotFound):
		writeAPIError(response, http.StatusNotFound, "product_not_found", "One of these products is not available.", nil, h.logger)
	default:
		h.logger.Error("comparison request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "comparison_unavailable", "Your comparison is temporarily unavailable.", nil, h.logger)
	}
}
