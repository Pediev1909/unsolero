package httpapi

import (
	"net/http"

	catalog "rigmark/internal/modules/catalog/domain"
)

func (h *Handler) adminDashboard(response http.ResponseWriter, request *http.Request) {
	result, err := h.admin.Dashboard(request.Context())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, dashboardDTO(result))
}

func (h *Handler) adminReferences(response http.ResponseWriter, request *http.Request) {
	result, err := h.admin.References(request.Context())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, referencesDTO(result))
}

func (h *Handler) adminListProducts(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListProducts(request.Context(), request.URL.Query().Get("q"), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, productPageDTO(result, page, pageSize))
}

func (h *Handler) adminGetProduct(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	result, err := h.admin.GetProduct(request.Context(), catalog.ProductID(id))
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminProductDTO(result))
}

func (h *Handler) adminCreateProduct(response http.ResponseWriter, request *http.Request) {
	var body productInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.admin.CreateProduct(request.Context(), principal.UserID, body.domain())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, adminProductDTO(result))
}

func (h *Handler) adminUpdateProduct(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	var body productInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.admin.UpdateProduct(request.Context(), principal.UserID, catalog.ProductID(id), body.domain())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminProductDTO(result))
}

func (h *Handler) adminProductStatus(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	if err := h.admin.SetProductStatus(request.Context(), principal.UserID, catalog.ProductID(id), catalog.ProductStatus(body.Status)); err != nil {
		h.writeAdminError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminListCategories(response http.ResponseWriter, request *http.Request) {
	items, err := h.admin.ListCategories(request.Context())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, categoriesDTO(items))
}

func (h *Handler) adminListBrands(response http.ResponseWriter, request *http.Request) {
	items, err := h.admin.ListBrands(request.Context())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, brandsDTO(items))
}

func (h *Handler) adminListMerchants(response http.ResponseWriter, request *http.Request) {
	items, err := h.admin.ListMerchants(request.Context())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, merchantsDTO(items))
}

func (h *Handler) adminListOffers(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListOffers(request.Context(), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, offersPageDTO(result, page, pageSize))
}

func (h *Handler) adminCreateOffer(response http.ResponseWriter, request *http.Request) {
	var body offerInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.admin.CreateOffer(request.Context(), principal.UserID, body.domain())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, offerDTOAdmin(result))
}

func (h *Handler) adminUpdateOffer(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("offerID"), h)
	if !ok {
		return
	}
	var body offerInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.admin.UpdateOffer(request.Context(), principal.UserID, id, body.domain())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, offerDTOAdmin(result))
}

func (h *Handler) adminListAffiliateLinks(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListAffiliateLinks(request.Context(), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, affiliatePageDTO(result, page, pageSize))
}

func (h *Handler) adminUpdateAffiliateLink(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("linkID"), h)
	if !ok {
		return
	}
	var body affiliateInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	result, err := h.admin.UpdateAffiliateLink(request.Context(), principal.UserID, id, body.domain())
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, affiliateDTO(result))
}

func (h *Handler) adminListRecommendations(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListRecommendations(request.Context(), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, recommendationsPageDTO(result, page, pageSize))
}

func (h *Handler) adminGetRecommendation(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("recommendationID"), h)
	if !ok {
		return
	}
	result, err := h.admin.GetRecommendation(request.Context(), id)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, recommendationDetailDTO(result))
}

func (h *Handler) adminListUsers(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListUsers(request.Context(), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, usersPageDTO(result, page, pageSize))
}

func (h *Handler) adminListEvents(response http.ResponseWriter, request *http.Request) {
	page, pageSize, ok := adminPagination(response, request, h)
	if !ok {
		return
	}
	result, err := h.admin.ListEvents(request.Context(), request.URL.Query().Get("name"), page, pageSize)
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, eventsPageDTO(result, page, pageSize))
}
