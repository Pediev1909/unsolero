package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	admin "rigmark/internal/modules/admin/application"
	admindomain "rigmark/internal/modules/admin/domain"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
)

const maximumAdminBodyBytes = 256 * 1024

type productInputRequest struct {
	CategoryID       string `json:"category_id"`
	BrandID          string `json:"brand_id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Description      string `json:"description"`
	PriceMinor       int64  `json:"price_minor"`
	Currency         string `json:"currency"`
	LengthMM         int64  `json:"length_mm"`
	WidthMM          int64  `json:"width_mm"`
	HeightMM         int64  `json:"height_mm"`
	WeightGrams      int64  `json:"weight_grams"`
	MaxCapacityGrams *int64 `json:"max_capacity_grams"`
	Material         string `json:"material"`
	WarrantyMonths   int16  `json:"warranty_months"`
	QualityScore     int16  `json:"quality_score"`
	ValueScore       int16  `json:"value_score"`
	DurabilityScore  int16  `json:"durability_score"`
	BeginnerScore    int16  `json:"beginner_score"`
	AdvancedScore    int16  `json:"advanced_score"`
	ApartmentScore   int16  `json:"apartment_score"`
	NoiseScore       int16  `json:"noise_score"`
	PortabilityScore int16  `json:"portability_score"`
}

type imageInputRequest struct {
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type attributeInputRequest struct {
	Type         string   `json:"type"`
	NumericValue *float64 `json:"numeric_value"`
	TextValue    *string  `json:"text_value"`
	BooleanValue *bool    `json:"boolean_value"`
	Unit         *string  `json:"unit"`
	IsFilterable bool     `json:"is_filterable"`
}

type affiliateInputRequest struct {
	Provider           string  `json:"provider"`
	DestinationURL     string  `json:"destination_url"`
	ExternalReference  *string `json:"external_reference"`
	DisclosureLabel    string  `json:"disclosure_label"`
	IsActive           bool    `json:"is_active"`
	Priority           int16   `json:"priority"`
	ProgramID          *string `json:"program_id"`
	CommissionType     string  `json:"commission_type"`
	CommissionRateBPS  *int    `json:"commission_rate_bps"`
	CommissionAmount   *int64  `json:"commission_amount_minor"`
	CommissionCurrency *string `json:"commission_currency"`
}

type offerInputRequest struct {
	MerchantID    string                 `json:"merchant_id"`
	ProductID     string                 `json:"product_id"`
	MerchantSKU   string                 `json:"merchant_sku"`
	ProductURL    string                 `json:"product_url"`
	PriceMinor    int64                  `json:"price_minor"`
	ShippingMinor int64                  `json:"shipping_minor"`
	Currency      string                 `json:"currency"`
	Availability  string                 `json:"availability"`
	Condition     string                 `json:"condition"`
	IsActive      bool                   `json:"is_active"`
	Affiliate     *affiliateInputRequest `json:"affiliate"`
}

func (h *Handler) decodeAdminJSON(response http.ResponseWriter, request *http.Request, destination any) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(response, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use application/json for this request.", nil, h.logger)
		return false
	}
	request.Body = http.MaxBytesReader(response, request.Body, maximumAdminBodyBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body is invalid.", nil, h.logger)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(response, http.StatusBadRequest, "invalid_json", "The request body must contain one JSON object.", nil, h.logger)
		return false
	}
	return true
}

func (h *Handler) writeAdminError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, admin.ErrInvalidInput):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_admin_input", "Check the submitted fields.", nil, h.logger)
	case errors.Is(err, adminports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "admin_entity_not_found", "The requested record was not found.", nil, h.logger)
	case errors.Is(err, adminports.ErrConflict):
		writeAPIError(response, http.StatusConflict, "admin_entity_conflict", "A record with these identifiers already exists.", nil, h.logger)
	default:
		h.logger.Error("admin request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "admin_unavailable", "Administration is temporarily unavailable.", nil, h.logger)
	}
}

func (h *Handler) writeAdminJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, status, value, h.logger)
}

func adminPagination(response http.ResponseWriter, request *http.Request, handler *Handler) (int, int, bool) {
	page, pageSize := 1, 30
	var err error
	if raw := request.URL.Query().Get("page"); raw != "" {
		page, err = strconv.Atoi(raw)
	}
	if err == nil {
		if raw := request.URL.Query().Get("page_size"); raw != "" {
			pageSize, err = strconv.Atoi(raw)
		}
	}
	if err != nil || page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_pagination", "Page values are invalid.", nil, handler.logger)
		return 0, 0, false
	}
	return page, pageSize, true
}

func adminUUIDPath(response http.ResponseWriter, value string, handler *Handler) (string, bool) {
	if !validUUID(strings.TrimSpace(value)) {
		writeAPIError(response, http.StatusNotFound, "admin_entity_not_found", "The requested record was not found.", nil, handler.logger)
		return "", false
	}
	return strings.TrimSpace(value), true
}

func (request productInputRequest) domain() admindomain.ProductInput {
	return admindomain.ProductInput{CategoryID: catalog.CategoryID(request.CategoryID), BrandID: catalog.BrandID(request.BrandID), Name: request.Name, Slug: request.Slug, Description: request.Description, Price: catalog.Money{AmountMinor: request.PriceMinor, Currency: request.Currency}, Dimensions: catalog.Dimensions{LengthMM: request.LengthMM, WidthMM: request.WidthMM, HeightMM: request.HeightMM}, WeightGrams: request.WeightGrams, MaxCapacityGrams: request.MaxCapacityGrams, Material: request.Material, WarrantyMonths: request.WarrantyMonths, Scores: catalog.Scores{Quality: request.QualityScore, Value: request.ValueScore, Durability: request.DurabilityScore, Beginner: request.BeginnerScore, Advanced: request.AdvancedScore, Apartment: request.ApartmentScore, Noise: request.NoiseScore, Portability: request.PortabilityScore}}
}

func (request affiliateInputRequest) domain() admindomain.AffiliateLinkInput {
	return admindomain.AffiliateLinkInput{Provider: request.Provider, DestinationURL: request.DestinationURL, ExternalReference: request.ExternalReference, DisclosureLabel: request.DisclosureLabel, IsActive: request.IsActive, Priority: request.Priority, ProgramID: request.ProgramID, CommissionType: request.CommissionType, CommissionRateBPS: request.CommissionRateBPS, CommissionAmount: request.CommissionAmount, CommissionCurrency: request.CommissionCurrency}
}

func (request offerInputRequest) domain() admindomain.OfferInput {
	result := admindomain.OfferInput{MerchantID: request.MerchantID, ProductID: request.ProductID, MerchantSKU: request.MerchantSKU, ProductURL: request.ProductURL, PriceMinor: request.PriceMinor, ShippingMinor: request.ShippingMinor, Currency: request.Currency, Availability: request.Availability, Condition: request.Condition, IsActive: request.IsActive}
	if request.Affiliate != nil {
		affiliate := request.Affiliate.domain()
		result.Affiliate = &affiliate
	}
	return result
}
