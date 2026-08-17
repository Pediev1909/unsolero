package httpapi

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"

	admindomain "rigmark/internal/modules/admin/domain"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

func (h *Handler) adminAddImage(response http.ResponseWriter, request *http.Request) {
	id, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	mediaType, _, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if mediaType == "multipart/form-data" {
		h.adminUploadImage(response, request, catalog.ProductID(id))
		return
	}
	var body imageInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	image, err := h.admin.AddImage(request.Context(), principal.UserID, catalog.ProductID(id), admindomain.ImageInput{URL: body.URL, AltText: body.AltText, SortOrder: body.SortOrder, IsPrimary: body.IsPrimary})
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, adminImageDTO(image))
}

func (h *Handler) adminUploadImage(response http.ResponseWriter, request *http.Request, productID catalog.ProductID) {
	request.Body = http.MaxBytesReader(response, request.Body, 6*1024*1024)
	if err := request.ParseMultipartForm(6 * 1024 * 1024); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_image_upload", "The image upload is invalid or exceeds 5 MB.", nil, h.logger)
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}
	file, header, err := request.FormFile("file")
	if err != nil || header.Size < 1 || header.Size > 5*1024*1024 {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_image_upload", "Choose a JPEG, PNG, or WebP image up to 5 MB.", nil, h.logger)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 5*1024*1024+1))
	if err != nil || len(data) > 5*1024*1024 {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_image_upload", "Choose a JPEG, PNG, or WebP image up to 5 MB.", nil, h.logger)
		return
	}
	sortOrder, err := strconv.Atoi(request.FormValue("sort_order"))
	if err != nil {
		sortOrder = 0
	}
	isPrimary, _ := strconv.ParseBool(request.FormValue("is_primary"))
	principal, _ := principalFromContext(request.Context())
	image, err := h.admin.UploadImage(request.Context(), principal.UserID, productID, admindomain.ImageUpload{Data: data, MIMEType: http.DetectContentType(data), AltText: request.FormValue("alt_text"), SortOrder: sortOrder, IsPrimary: isPrimary})
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusCreated, adminImageDTO(image))
}

func (h *Handler) productImage(response http.ResponseWriter, request *http.Request) {
	data, contentType, err := h.admin.OpenImage(request.Context(), request.PathValue("file"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, adminports.ErrNotFound) {
			http.NotFound(response, request)
			return
		}
		h.logger.Error("serve product image", "error", err)
		http.Error(response, "Image unavailable", http.StatusServiceUnavailable)
		return
	}
	response.Header().Set("Content-Type", contentType)
	response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(data)
}

func (h *Handler) adminDeleteImage(response http.ResponseWriter, request *http.Request) {
	productID, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	imageID, ok := adminUUIDPath(response, request.PathValue("imageID"), h)
	if !ok {
		return
	}
	principal, _ := principalFromContext(request.Context())
	if err := h.admin.DeleteImage(request.Context(), principal.UserID, catalog.ProductID(productID), imageID); err != nil {
		h.writeAdminError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) adminUpsertAttribute(response http.ResponseWriter, request *http.Request) {
	productID, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	var body attributeInputRequest
	if !h.decodeAdminJSON(response, request, &body) {
		return
	}
	principal, _ := principalFromContext(request.Context())
	attribute, err := h.admin.UpsertAttribute(request.Context(), principal.UserID, catalog.ProductID(productID), admindomain.AttributeInput{Key: request.PathValue("key"), Type: catalog.AttributeType(body.Type), NumericValue: body.NumericValue, TextValue: body.TextValue, BooleanValue: body.BooleanValue, Unit: body.Unit, IsFilterable: body.IsFilterable})
	if err != nil {
		h.writeAdminError(response, err)
		return
	}
	h.writeAdminJSON(response, http.StatusOK, adminAttributeDTO(attribute))
}

func (h *Handler) adminDeleteAttribute(response http.ResponseWriter, request *http.Request) {
	productID, ok := adminUUIDPath(response, request.PathValue("productID"), h)
	if !ok {
		return
	}
	principal, _ := principalFromContext(request.Context())
	if err := h.admin.DeleteAttribute(request.Context(), principal.UserID, catalog.ProductID(productID), request.PathValue("key")); err != nil {
		h.writeAdminError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
