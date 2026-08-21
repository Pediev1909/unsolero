package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/application"
	"rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
	commercedomain "rigmark/internal/modules/commerce/domain"
	planning "rigmark/internal/modules/planning/application"
	planningports "rigmark/internal/modules/planning/ports"
)

type imageResponse struct {
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	IsPrimary bool   `json:"is_primary"`
	WidthPX   *int   `json:"width_px,omitempty"`
	HeightPX  *int   `json:"height_px,omitempty"`
}

type namedResourceResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type moneyResponse struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

type scoresResponse struct {
	Quality     int16 `json:"quality"`
	Value       int16 `json:"value"`
	Durability  int16 `json:"durability"`
	Beginner    int16 `json:"beginner"`
	Advanced    int16 `json:"advanced"`
	Apartment   int16 `json:"apartment"`
	Noise       int16 `json:"noise"`
	Portability int16 `json:"portability"`
}

type insightResponse struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Score int16  `json:"score"`
}

type keySpecificationResponse struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type productSummaryResponse struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Slug             string                   `json:"slug"`
	Brand            namedResourceResponse    `json:"brand"`
	Category         namedResourceResponse    `json:"category"`
	Price            moneyResponse            `json:"price"`
	PrimaryImage     *imageResponse           `json:"primary_image"`
	KeySpecification keySpecificationResponse `json:"key_specification"`
	Suitability      []insightResponse        `json:"suitability"`
	Scores           scoresResponse           `json:"scores"`
	IsDemo           bool                     `json:"is_demo"`
}

type attributeResponse struct {
	Key          string   `json:"key"`
	Type         string   `json:"type"`
	NumericValue *float64 `json:"numeric_value,omitempty"`
	TextValue    *string  `json:"text_value,omitempty"`
	BooleanValue *bool    `json:"boolean_value,omitempty"`
	Unit         *string  `json:"unit,omitempty"`
}

type dimensionsResponse struct {
	LengthMM int64 `json:"length_mm"`
	WidthMM  int64 `json:"width_mm"`
	HeightMM int64 `json:"height_mm"`
}

type productDetailResponse struct {
	productSummaryResponse
	Description      string                    `json:"description"`
	Images           []imageResponse           `json:"images"`
	Dimensions       dimensionsResponse        `json:"dimensions"`
	WeightGrams      int64                     `json:"weight_grams"`
	MaxCapacityGrams *int64                    `json:"max_capacity_grams"`
	Material         string                    `json:"material"`
	WarrantyMonths   int16                     `json:"warranty_months"`
	Attributes       []attributeResponse       `json:"attributes"`
	Strengths        []insightResponse         `json:"strengths"`
	Weaknesses       []insightResponse         `json:"weaknesses"`
	UseCases         []insightResponse         `json:"use_cases"`
	Alternatives     []productSummaryResponse  `json:"alternatives"`
	Evidence         []productEvidenceResponse `json:"evidence"`
	FactRevisionID   string                    `json:"fact_revision_id"`
	ScoreRevisionID  string                    `json:"score_revision_id"`
}

type productEvidenceResponse struct {
	FactKey        string     `json:"fact_key"`
	Classification string     `json:"classification"`
	SourceType     string     `json:"source_type"`
	SourceTitle    string     `json:"source_title"`
	SourceURL      *string    `json:"source_url"`
	ObservedAt     time.Time  `json:"observed_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	Confidence     int16      `json:"confidence"`
	IsFictional    bool       `json:"is_fictional"`
}

type productPageResponse struct {
	Products   []productSummaryResponse `json:"products"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Total      int                      `json:"total"`
	TotalPages int                      `json:"total_pages"`
}

type categoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	// PublishedProducts lets a listing say "6 tools" beside a category
	// instead of making the reader open it to find out, and lets a category
	// with nothing in it be left out of an index rather than shown empty.
	PublishedProducts int    `json:"published_products"`
	Description       string `json:"description"`
}

type brandResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	PublishedProducts int     `json:"published_products"`
	Description       string  `json:"description"`
	CountryCode       *string `json:"country_code,omitempty"`
}

func (h *Handler) listProducts(response http.ResponseWriter, request *http.Request) {
	query, err := catalogQuery(request)
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	page, err := h.catalog.Search(request.Context(), query)
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	products := make([]productSummaryResponse, 0, len(page.Products))
	for _, product := range page.Products {
		products = append(products, productSummaryDTO(product))
	}
	response.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(response, http.StatusOK, productPageResponse{
		Products: products, Page: page.Page, PageSize: page.PageSize,
		Total: page.Total, TotalPages: page.TotalPages,
	}, h.logger)
}

func (h *Handler) getProduct(response http.ResponseWriter, request *http.Request) {
	detail, err := h.catalog.GetProduct(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "public, max-age=60")
	writeJSON(response, http.StatusOK, productDetailDTO(detail), h.logger)
}

func (h *Handler) listCategories(response http.ResponseWriter, request *http.Request) {
	categories, err := h.catalog.ListCategories(request.Context())
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	result := make([]categoryResponse, 0, len(categories))
	for _, category := range categories {
		result = append(result, categoryDTO(category))
	}
	response.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(response, http.StatusOK, result, h.logger)
}

func (h *Handler) getCategory(response http.ResponseWriter, request *http.Request) {
	category, err := h.catalog.GetCategory(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(response, http.StatusOK, categoryDTO(category), h.logger)
}

func (h *Handler) listBrands(response http.ResponseWriter, request *http.Request) {
	brands, err := h.catalog.ListBrands(request.Context())
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	result := make([]brandResponse, 0, len(brands))
	for _, brand := range brands {
		result = append(result, brandDTO(brand))
	}
	response.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(response, http.StatusOK, result, h.logger)
}

func (h *Handler) getBrand(response http.ResponseWriter, request *http.Request) {
	brand, err := h.catalog.GetBrand(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(response, http.StatusOK, brandDTO(brand), h.logger)
}

func catalogQuery(request *http.Request) (catalog.Query, error) {
	values := request.URL.Query()
	query := catalog.Query{
		Search: values.Get("q"), CategorySlug: values.Get("category"),
		BrandSlug: values.Get("brand"), Sort: values.Get("sort"),
	}
	if rawIDs := values.Get("ids"); rawIDs != "" {
		parts := strings.Split(rawIDs, ",")
		if len(parts) > 48 {
			return catalog.Query{}, catalog.ErrInvalidQuery
		}
		for _, value := range parts {
			value = strings.TrimSpace(value)
			if !validUUID(value) {
				return catalog.Query{}, catalog.ErrInvalidQuery
			}
			query.ProductIDs = append(query.ProductIDs, domain.ProductID(value))
		}
	}
	var err error
	if query.Page, err = optionalPositiveInt(values.Get("page"), 1); err != nil {
		return catalog.Query{}, catalog.ErrInvalidQuery
	}
	if query.PageSize, err = optionalPositiveInt(values.Get("page_size"), 12); err != nil {
		return catalog.Query{}, catalog.ErrInvalidQuery
	}
	if query.MinPriceMinor, err = optionalNonNegativeInt64(values.Get("min_price_minor")); err != nil {
		return catalog.Query{}, catalog.ErrInvalidQuery
	}
	if query.MaxPriceMinor, err = optionalNonNegativeInt64(values.Get("max_price_minor")); err != nil {
		return catalog.Query{}, catalog.ErrInvalidQuery
	}
	return query, nil
}

func optionalPositiveInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, catalog.ErrInvalidQuery
	}
	return parsed, nil
}

func optionalNonNegativeInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return nil, catalog.ErrInvalidQuery
	}
	return &parsed, nil
}

func (h *Handler) writeCatalogError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ports.ErrNotFound):
		writeAPIError(response, http.StatusNotFound, "catalog_not_found", "This catalog page could not be found.", nil, h.logger)
	case errors.Is(err, catalog.ErrInvalidQuery):
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_catalog_query", "Check the catalog filters and try again.", nil, h.logger)
	default:
		h.logger.Error("catalog request failed", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "catalog_unavailable", "The catalog is temporarily unavailable.", nil, h.logger)
	}
}

func productSummaryDTO(product domain.Product) productSummaryResponse {
	var primaryImage *imageResponse
	for _, image := range product.Images {
		if image.IsPrimary || primaryImage == nil {
			value := imageDTO(image)
			primaryImage = &value
		}
		if image.IsPrimary {
			break
		}
	}
	return productSummaryResponse{
		ID: string(product.ID), Name: product.Name, Slug: product.Slug,
		Brand:        namedResourceResponse{Name: product.BrandName, Slug: product.BrandSlug},
		Category:     namedResourceResponse{Name: product.CategoryName, Slug: product.CategorySlug},
		Price:        moneyResponse{AmountMinor: product.Price.AmountMinor, Currency: product.Price.Currency},
		PrimaryImage: primaryImage, KeySpecification: keySpecificationDTO(product),
		Suitability: insightDTOs(product.Suitability()), Scores: scoreDTO(product.Scores),
		IsDemo: strings.HasPrefix(product.Slug, "demo-"),
	}
}

func productDetailDTO(detail catalog.ProductDetail) productDetailResponse {
	product := detail.Product
	images := make([]imageResponse, 0, len(product.Images))
	for _, image := range product.Images {
		images = append(images, imageDTO(image))
	}
	attributes := make([]attributeResponse, 0, len(product.Attributes))
	for _, attribute := range product.Attributes {
		attributes = append(attributes, attributeResponse{
			Key: attribute.Key, Type: string(attribute.Type), NumericValue: attribute.NumericValue,
			TextValue: attribute.TextValue, BooleanValue: attribute.BooleanValue, Unit: attribute.Unit,
		})
	}
	alternatives := make([]productSummaryResponse, 0, len(detail.Alternatives))
	for _, alternative := range detail.Alternatives {
		alternatives = append(alternatives, productSummaryDTO(alternative))
	}
	evidence := make([]productEvidenceResponse, 0, len(product.Evidence))
	for _, item := range product.Evidence {
		evidence = append(evidence, productEvidenceResponse{
			FactKey: item.FactKey, Classification: item.Classification,
			SourceType: item.SourceType, SourceTitle: item.SourceTitle,
			SourceURL: item.SourceURL, ObservedAt: item.ObservedAt,
			ExpiresAt: item.ExpiresAt, Confidence: item.Confidence,
			IsFictional: item.IsFictional,
		})
	}
	return productDetailResponse{
		productSummaryResponse: productSummaryDTO(product), Description: product.Description,
		Images: images, Dimensions: dimensionsResponse{
			LengthMM: product.Dimensions.LengthMM, WidthMM: product.Dimensions.WidthMM,
			HeightMM: product.Dimensions.HeightMM,
		}, WeightGrams: product.WeightGrams, MaxCapacityGrams: product.MaxCapacityGrams,
		Material: product.Material, WarrantyMonths: product.WarrantyMonths, Attributes: attributes,
		Strengths: insightDTOs(detail.Strengths), Weaknesses: insightDTOs(detail.Considerations),
		UseCases: insightDTOs(detail.UseCases), Alternatives: alternatives,
		Evidence: evidence, FactRevisionID: product.FactRevisionID,
		ScoreRevisionID: product.ScoreRevisionID,
	}
}

func keySpecificationDTO(product domain.Product) keySpecificationResponse {
	if product.MaxCapacityGrams != nil {
		return keySpecificationResponse{Label: "Maximum capacity", Value: formatKilograms(*product.MaxCapacityGrams)}
	}
	if len(product.Attributes) > 0 {
		attribute := product.Attributes[0]
		label := strings.ReplaceAll(attribute.Key, "_", " ")
		value := "—"
		switch {
		case attribute.NumericValue != nil:
			value = strconv.FormatFloat(*attribute.NumericValue, 'f', -1, 64)
			if attribute.Unit != nil {
				value += " " + *attribute.Unit
			}
		case attribute.TextValue != nil:
			value = strings.ReplaceAll(*attribute.TextValue, "_", " ")
		case attribute.BooleanValue != nil && *attribute.BooleanValue:
			value = "Yes"
		case attribute.BooleanValue != nil:
			value = "No"
		}
		return keySpecificationResponse{Label: label, Value: value}
	}
	if !product.IsPhysical {
		// A subscription has no headline physical measurement. Falling through
		// to weight reported "0 kg", which reads as broken data rather than as
		// an absent one.
		return keySpecificationResponse{Label: "Billing", Value: "Per month"}
	}
	return keySpecificationResponse{Label: "Product weight", Value: formatKilograms(product.WeightGrams)}
}

func formatKilograms(grams int64) string {
	return strconv.FormatFloat(float64(grams)/1000, 'f', -1, 64) + " kg"
}

func insightDTOs(insights []domain.SuitabilityInsight) []insightResponse {
	result := make([]insightResponse, 0, len(insights))
	for _, insight := range insights {
		result = append(result, insightResponse{Key: insight.Key, Label: insight.Label, Score: insight.Score})
	}
	return result
}

func imageDTO(image domain.ProductImage) imageResponse {
	return imageResponse{
		URL: image.URL, AltText: image.AltText, IsPrimary: image.IsPrimary,
		WidthPX: image.WidthPX, HeightPX: image.HeightPX,
	}
}

func scoreDTO(scores domain.Scores) scoresResponse {
	return scoresResponse{
		Quality: scores.Quality, Value: scores.Value, Durability: scores.Durability,
		Beginner: scores.Beginner, Advanced: scores.Advanced, Apartment: scores.Apartment,
		Noise: scores.Noise, Portability: scores.Portability,
	}
}

func categoryDTO(category domain.Category) categoryResponse {
	return categoryResponse{
		ID: string(category.ID), Name: category.Name, Slug: category.Slug,
		PublishedProducts: category.PublishedProducts,
		Description:       category.Description,
	}
}

func brandDTO(brand domain.Brand) brandResponse {
	return brandResponse{
		ID: string(brand.ID), Name: brand.Name, Slug: brand.Slug,
		PublishedProducts: brand.PublishedProducts,
		Description:       brand.Description, CountryCode: brand.CountryCode,
	}
}

type offerResponse struct {
	ID               string                `json:"id"`
	Merchant         offerMerchantResponse `json:"merchant"`
	Price            moneyResponse         `json:"price"`
	ShippingMinor    int64                 `json:"shipping_minor"`
	LandedPriceMinor int64                 `json:"landed_price_minor"`
	Availability     string                `json:"availability"`
	Condition        string                `json:"condition"`
	LastCheckedAt    string                `json:"last_checked_at"`
	ObservedAt       *string               `json:"observed_at"`
	ExpiresAt        *string               `json:"expires_at"`
	FreshnessStatus  string                `json:"freshness_status"`
	PurchasePath     *string               `json:"purchase_path"`
	DisclosureLabel  *string               `json:"disclosure_label"`
}

type offerMerchantResponse struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	CountryCode string `json:"country_code"`
	TrustScore  int16  `json:"trust_score"`
}

func (h *Handler) listOffers(response http.ResponseWriter, request *http.Request) {
	detail, err := h.catalog.GetProduct(request.Context(), request.PathValue("slug"))
	if err != nil {
		h.writeCatalogError(response, err)
		return
	}
	offers, err := h.commerce.ListOffers(request.Context(), detail.Product.ID, detail.Product.Price.Currency)
	if err != nil {
		h.logger.Error("list product offers", "error", err)
		writeAPIError(response, http.StatusInternalServerError, "offers_unavailable", "Merchant offers are temporarily unavailable.", nil, h.logger)
		return
	}
	result := make([]offerResponse, 0, len(offers))
	for _, offer := range offers {
		result = append(result, offerDTO(offer))
	}
	response.Header().Set("Cache-Control", "public, max-age=30")
	writeJSON(response, http.StatusOK, result, h.logger)
}

func offerDTO(offer commercedomain.Offer) offerResponse {
	var purchasePath *string
	var disclosure *string
	if len(offer.AffiliateLinks) > 0 {
		path := "/api/affiliate/click/" + string(offer.ID)
		purchasePath = &path
		disclosure = &offer.AffiliateLinks[0].DisclosureLabel
	}
	var observedAt *string
	if offer.ProviderObservedAt != nil {
		value := offer.ProviderObservedAt.UTC().Format(time.RFC3339)
		observedAt = &value
	}
	var expiresAt *string
	if offer.ExpiresAt != nil {
		value := offer.ExpiresAt.UTC().Format(time.RFC3339)
		expiresAt = &value
	}
	return offerResponse{
		ID: string(offer.ID), Merchant: offerMerchantResponse{
			Name: offer.Merchant.Name, Slug: offer.Merchant.Slug,
			CountryCode: offer.Merchant.CountryCode, TrustScore: offer.Merchant.TrustScore,
		}, Price: moneyResponse{AmountMinor: offer.Price.AmountMinor, Currency: offer.Price.Currency},
		ShippingMinor: offer.ShippingMinor, LandedPriceMinor: offer.LandedPriceMinor(),
		Availability: offer.Availability, Condition: offer.Condition,
		LastCheckedAt: offer.LastCheckedAt.UTC().Format("2006-01-02T15:04:05Z"),
		ObservedAt:    observedAt, ExpiresAt: expiresAt, FreshnessStatus: "fresh",
		PurchasePath: purchasePath, DisclosureLabel: disclosure,
	}
}

type wishlistResponse struct {
	ProductIDs []string `json:"product_ids"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	Total      int      `json:"total"`
	TotalPages int      `json:"total_pages"`
}

func (h *Handler) listWishlist(response http.ResponseWriter, request *http.Request) {
	page, err := optionalPositiveInt(request.URL.Query().Get("page"), 1)
	if err != nil {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_pagination", "Wishlist pagination is invalid.", nil, h.logger)
		return
	}
	pageSize, err := optionalPositiveInt(request.URL.Query().Get("page_size"), 50)
	if err != nil || page > 10_000 || pageSize > 100 {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_pagination", "Wishlist pagination is invalid.", nil, h.logger)
		return
	}
	principal, _ := principalFromContext(request.Context())
	wishlistPage, err := h.wishlist.List(request.Context(), principal.UserID, page, pageSize)
	if err != nil {
		h.writeWishlistError(response, err)
		return
	}
	result := make([]string, 0, len(wishlistPage.ProductIDs))
	for _, productID := range wishlistPage.ProductIDs {
		result = append(result, string(productID))
	}
	totalPages := 0
	if wishlistPage.Total > 0 {
		totalPages = (wishlistPage.Total + pageSize - 1) / pageSize
	}
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, wishlistResponse{ProductIDs: result, Page: page, PageSize: pageSize,
		Total: wishlistPage.Total, TotalPages: totalPages}, h.logger)
}

func (h *Handler) saveWishlist(response http.ResponseWriter, request *http.Request) {
	productID := request.PathValue("productID")
	if !validUUID(productID) {
		h.writeWishlistError(response, planningports.ErrProductNotFound)
		return
	}
	principal, _ := principalFromContext(request.Context())
	if err := h.wishlist.Save(request.Context(), principal.UserID, domain.ProductID(productID)); err != nil {
		h.writeWishlistError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteWishlist(response http.ResponseWriter, request *http.Request) {
	productID := request.PathValue("productID")
	if !validUUID(productID) {
		h.writeWishlistError(response, planningports.ErrProductNotFound)
		return
	}
	principal, _ := principalFromContext(request.Context())
	if err := h.wishlist.Delete(request.Context(), principal.UserID, domain.ProductID(productID)); err != nil {
		h.writeWishlistError(response, err)
		return
	}
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeWishlistError(response http.ResponseWriter, err error) {
	if errors.Is(err, planning.ErrInvalidPagination) {
		writeAPIError(response, http.StatusUnprocessableEntity, "invalid_pagination", "Wishlist pagination is invalid.", nil, h.logger)
		return
	}
	if errors.Is(err, planningports.ErrProductNotFound) {
		writeAPIError(response, http.StatusNotFound, "product_not_found", "This product is not available.", nil, h.logger)
		return
	}
	h.logger.Error("wishlist request failed", "error", err)
	writeAPIError(response, http.StatusInternalServerError, "wishlist_unavailable", "Saved products are temporarily unavailable.", nil, h.logger)
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for index, character := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !strings.ContainsRune("0123456789abcdefABCDEF", character) {
			return false
		}
	}
	return true
}
