package httpapi

import (
	"encoding/json"
	"time"

	admin "rigmark/internal/modules/admin/domain"
	catalog "rigmark/internal/modules/catalog/domain"
)

type adminPageResponse[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type adminDashboardResponse struct {
	Counts struct {
		Products        int64 `json:"products"`
		Published       int64 `json:"published"`
		Offers          int64 `json:"offers"`
		ActiveOffers    int64 `json:"active_offers"`
		Users           int64 `json:"users"`
		Recommendations int64 `json:"recommendations"`
	} `json:"counts"`
	Analytics struct {
		RecommendationStarts     int64 `json:"recommendation_starts"`
		CompletedRecommendations int64 `json:"completed_recommendations"`
		ProductViews             int64 `json:"product_views"`
		AffiliateClicks          int64 `json:"affiliate_clicks"`
		SavedProducts            int64 `json:"saved_products"`
		SavedSetups              int64 `json:"saved_setups"`
	} `json:"analytics"`
	Readiness adminReadinessResponse `json:"readiness"`
}

// The dashboard reports monetization readiness so an operator can see, without
// opening another page, how much of the published catalog can actually earn.
type adminReadinessResponse struct {
	PublishedProducts    int64                  `json:"published_products"`
	WithoutActiveOffer   int64                  `json:"without_active_offer"`
	WithoutAffiliateLink int64                  `json:"without_affiliate_link"`
	EarningReady         int64                  `json:"earning_ready"`
	CommerceProviders    int64                  `json:"commerce_providers"`
	PublishedContent     int64                  `json:"published_content"`
	Blocked              []adminBlockedResponse `json:"blocked"`
}

type adminBlockedResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Reason string `json:"reason"`
}

type adminProductResponse struct {
	ID               string                   `json:"id"`
	CategoryID       string                   `json:"category_id"`
	CategoryName     string                   `json:"category_name"`
	BrandID          string                   `json:"brand_id"`
	BrandName        string                   `json:"brand_name"`
	Name             string                   `json:"name"`
	Slug             string                   `json:"slug"`
	Description      string                   `json:"description"`
	PriceMinor       int64                    `json:"price_minor"`
	Currency         string                   `json:"currency"`
	LengthMM         int64                    `json:"length_mm"`
	WidthMM          int64                    `json:"width_mm"`
	HeightMM         int64                    `json:"height_mm"`
	WeightGrams      int64                    `json:"weight_grams"`
	MaxCapacityGrams *int64                   `json:"max_capacity_grams"`
	Material         string                   `json:"material"`
	WarrantyMonths   int16                    `json:"warranty_months"`
	Scores           scoresResponse           `json:"scores"`
	Status           string                   `json:"status"`
	Images           []adminImageResponse     `json:"images"`
	Attributes       []adminAttributeResponse `json:"attributes"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type adminImageResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

type adminAttributeResponse struct {
	ID           string   `json:"id"`
	Key          string   `json:"key"`
	Type         string   `json:"type"`
	NumericValue *float64 `json:"numeric_value"`
	TextValue    *string  `json:"text_value"`
	BooleanValue *bool    `json:"boolean_value"`
	Unit         *string  `json:"unit"`
	IsFilterable bool     `json:"is_filterable"`
}

type categoryAdminResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	Products  int64     `json:"products"`
	UpdatedAt time.Time `json:"updated_at"`
}
type brandAdminResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	Products  int64     `json:"products"`
	UpdatedAt time.Time `json:"updated_at"`
}
type merchantAdminResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	WebsiteURL  string    `json:"website_url"`
	CountryCode string    `json:"country_code"`
	TrustScore  int16     `json:"trust_score"`
	Status      string    `json:"status"`
	Offers      int64     `json:"offers"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type productReferenceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
type adminReferencesResponse struct {
	Categories []categoryAdminResponse    `json:"categories"`
	Brands     []brandAdminResponse       `json:"brands"`
	Merchants  []merchantAdminResponse    `json:"merchants"`
	Products   []productReferenceResponse `json:"products"`
}
type offerAdminResponse struct {
	ID              string     `json:"id"`
	MerchantID      string     `json:"merchant_id"`
	MerchantName    string     `json:"merchant_name"`
	ProductID       string     `json:"product_id"`
	ProductName     string     `json:"product_name"`
	MerchantSKU     string     `json:"merchant_sku"`
	ProductURL      string     `json:"product_url"`
	PriceMinor      int64      `json:"price_minor"`
	ShippingMinor   int64      `json:"shipping_minor"`
	Currency        string     `json:"currency"`
	Availability    string     `json:"availability"`
	Condition       string     `json:"condition"`
	IsActive        bool       `json:"is_active"`
	LastCheckedAt   time.Time  `json:"last_checked_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	FreshnessStatus string     `json:"freshness_status"`
	AffiliateLinks  int64      `json:"affiliate_links"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
type affiliateAdminResponse struct {
	ID                 string    `json:"id"`
	OfferID            string    `json:"offer_id"`
	ProductName        string    `json:"product_name"`
	MerchantName       string    `json:"merchant_name"`
	Provider           string    `json:"provider"`
	DestinationURL     string    `json:"destination_url"`
	ExternalReference  *string   `json:"external_reference"`
	DisclosureLabel    string    `json:"disclosure_label"`
	IsActive           bool      `json:"is_active"`
	Priority           int16     `json:"priority"`
	ProgramID          *string   `json:"program_id"`
	CommissionType     string    `json:"commission_type"`
	CommissionRateBPS  *int      `json:"commission_rate_bps"`
	CommissionAmount   *int64    `json:"commission_amount_minor"`
	CommissionCurrency *string   `json:"commission_currency"`
	UpdatedAt          time.Time `json:"updated_at"`
}
type recommendationAdminResponse struct {
	ID              string    `json:"id"`
	SessionID       string    `json:"session_id"`
	UserEmail       *string   `json:"user_email"`
	Goal            string    `json:"goal"`
	Experience      string    `json:"experience"`
	SessionStatus   string    `json:"session_status"`
	ObjectiveScore  int16     `json:"objective_score"`
	TotalPriceMinor int64     `json:"total_price_minor"`
	Currency        string    `json:"currency"`
	PolicyVersion   string    `json:"policy_version"`
	EngineVersion   string    `json:"engine_version"`
	CreatedAt       time.Time `json:"created_at"`
}
type recommendationReasonResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Dimension string `json:"dimension"`
	Score     int16  `json:"score"`
}
type recommendationItemAdminResponse struct {
	ProductID      string                         `json:"product_id"`
	ProductName    string                         `json:"product_name"`
	ItemType       string                         `json:"item_type"`
	Rank           int16                          `json:"rank"`
	Quantity       int16                          `json:"quantity"`
	ObjectiveScore int16                          `json:"objective_score"`
	ReasonCode     string                         `json:"reason_code"`
	ReasonSummary  string                         `json:"reason_summary"`
	RejectionCode  *string                        `json:"rejection_code"`
	Reasons        []recommendationReasonResponse `json:"reasons"`
}
type adminScoreBreakdownResponse struct {
	Goal          int16 `json:"goal"`
	Budget        int16 `json:"budget"`
	Space         int16 `json:"space"`
	Experience    int16 `json:"experience"`
	Preference    int16 `json:"preference"`
	Quality       int16 `json:"quality"`
	Value         int16 `json:"value"`
	Durability    int16 `json:"durability"`
	Compatibility int16 `json:"compatibility"`
	Portability   int16 `json:"portability"`
	Noise         int16 `json:"noise"`
}
type recommendationDetailAdminResponse struct {
	Recommendation recommendationAdminResponse       `json:"recommendation"`
	Scores         adminScoreBreakdownResponse       `json:"scores"`
	Items          []recommendationItemAdminResponse `json:"items"`
}
type userAdminResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
type eventAdminResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	UserID        *string         `json:"user_id"`
	AnonymousID   *string         `json:"anonymous_id"`
	SessionID     *string         `json:"session_id"`
	Surface       string          `json:"surface"`
	Properties    json.RawMessage `json:"properties"`
	PagePath      *string         `json:"page_path"`
	TrafficSource *string         `json:"traffic_source"`
	TrafficMedium *string         `json:"traffic_medium"`
	Campaign      *string         `json:"campaign"`
	ReferrerHost  *string         `json:"referrer_host"`
	ConsentState  string          `json:"consent_state"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func dashboardDTO(value admin.Dashboard) adminDashboardResponse {
	var result adminDashboardResponse
	result.Counts.Products = value.Counts.Products
	result.Counts.Published = value.Counts.Published
	result.Counts.Offers = value.Counts.Offers
	result.Counts.ActiveOffers = value.Counts.ActiveOffers
	result.Counts.Users = value.Counts.Users
	result.Counts.Recommendations = value.Counts.Recommendations
	result.Analytics.RecommendationStarts = value.Analytics.RecommendationStarts
	result.Analytics.CompletedRecommendations = value.Analytics.CompletedRecommendations
	result.Analytics.ProductViews = value.Analytics.ProductViews
	result.Analytics.AffiliateClicks = value.Analytics.AffiliateClicks
	result.Analytics.SavedProducts = value.Analytics.SavedProducts
	result.Analytics.SavedSetups = value.Analytics.SavedSetups
	result.Readiness = readinessDTO(value.Readiness)
	return result
}

func readinessDTO(value admin.Readiness) adminReadinessResponse {
	blocked := make([]adminBlockedResponse, 0, len(value.Blocked))
	for _, item := range value.Blocked {
		blocked = append(blocked, adminBlockedResponse{
			ID: string(item.ID), Name: item.Name, Slug: item.Slug,
			Reason: string(item.Reason),
		})
	}
	return adminReadinessResponse{
		PublishedProducts:    value.PublishedProducts,
		WithoutActiveOffer:   value.WithoutActiveOffer,
		WithoutAffiliateLink: value.WithoutAffiliateLink,
		EarningReady:         value.EarningReady,
		CommerceProviders:    value.CommerceProviders,
		PublishedContent:     value.PublishedContent,
		Blocked:              blocked,
	}
}
func adminProductDTO(value catalog.Product) adminProductResponse {
	images := make([]adminImageResponse, 0, len(value.Images))
	for _, item := range value.Images {
		images = append(images, adminImageDTO(item))
	}
	attributes := make([]adminAttributeResponse, 0, len(value.Attributes))
	for _, item := range value.Attributes {
		attributes = append(attributes, adminAttributeDTO(item))
	}
	return adminProductResponse{ID: string(value.ID), CategoryID: string(value.CategoryID), CategoryName: value.CategoryName, BrandID: string(value.BrandID), BrandName: value.BrandName, Name: value.Name, Slug: value.Slug, Description: value.Description, PriceMinor: value.Price.AmountMinor, Currency: value.Price.Currency, LengthMM: value.Dimensions.LengthMM, WidthMM: value.Dimensions.WidthMM, HeightMM: value.Dimensions.HeightMM, WeightGrams: value.WeightGrams, MaxCapacityGrams: value.MaxCapacityGrams, Material: value.Material, WarrantyMonths: value.WarrantyMonths, Scores: scoreDTO(value.Scores), Status: string(value.Status), Images: images, Attributes: attributes, UpdatedAt: value.UpdatedAt}
}
func adminImageDTO(value catalog.ProductImage) adminImageResponse {
	return adminImageResponse{ID: value.ID, URL: value.URL, AltText: value.AltText, SortOrder: value.SortOrder, IsPrimary: value.IsPrimary}
}
func adminAttributeDTO(value catalog.Attribute) adminAttributeResponse {
	return adminAttributeResponse{ID: value.ID, Key: value.Key, Type: string(value.Type), NumericValue: value.NumericValue, TextValue: value.TextValue, BooleanValue: value.BooleanValue, Unit: value.Unit, IsFilterable: value.IsFilterable}
}
func productPageDTO(value admin.ProductPage, page, pageSize int) adminPageResponse[adminProductResponse] {
	items := make([]adminProductResponse, 0, len(value.Items))
	for _, item := range value.Items {
		items = append(items, adminProductDTO(item))
	}
	return pageDTO(items, value.Total, page, pageSize)
}
func categoriesDTO(values []admin.Category) []categoryAdminResponse {
	result := make([]categoryAdminResponse, 0, len(values))
	for _, v := range values {
		result = append(result, categoryAdminResponse{v.ID, v.Name, v.Slug, v.IsActive, v.Products, v.UpdatedAt})
	}
	return result
}
func brandsDTO(values []admin.Brand) []brandAdminResponse {
	result := make([]brandAdminResponse, 0, len(values))
	for _, v := range values {
		result = append(result, brandAdminResponse{v.ID, v.Name, v.Slug, v.IsActive, v.Products, v.UpdatedAt})
	}
	return result
}
func merchantsDTO(values []admin.Merchant) []merchantAdminResponse {
	result := make([]merchantAdminResponse, 0, len(values))
	for _, v := range values {
		result = append(result, merchantAdminResponse{v.ID, v.Name, v.Slug, v.WebsiteURL, v.CountryCode, v.TrustScore, v.Status, v.Offers, v.UpdatedAt})
	}
	return result
}
func referencesDTO(value admin.References) adminReferencesResponse {
	products := make([]productReferenceResponse, 0, len(value.Products))
	for _, v := range value.Products {
		products = append(products, productReferenceResponse{v.ID, v.Name, v.Slug})
	}
	return adminReferencesResponse{categoriesDTO(value.Categories), brandsDTO(value.Brands), merchantsDTO(value.Merchants), products}
}
func offerDTOAdmin(v admin.Offer) offerAdminResponse {
	return offerAdminResponse{v.ID, v.MerchantID, v.MerchantName, v.ProductID, v.ProductName, v.MerchantSKU, v.ProductURL, v.PriceMinor, v.ShippingMinor, v.Currency, v.Availability, v.Condition, v.IsActive, v.LastCheckedAt, v.ExpiresAt, v.FreshnessStatus, v.AffiliateLinks, v.UpdatedAt}
}
func offersPageDTO(v admin.Page[admin.Offer], page, pageSize int) adminPageResponse[offerAdminResponse] {
	items := make([]offerAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, offerDTOAdmin(item))
	}
	return pageDTO(items, v.Total, page, pageSize)
}
func affiliateDTO(v admin.AffiliateLink) affiliateAdminResponse {
	return affiliateAdminResponse{v.ID, v.OfferID, v.ProductName, v.MerchantName, v.Provider, v.DestinationURL, v.ExternalReference, v.DisclosureLabel, v.IsActive, v.Priority, v.ProgramID, v.CommissionType, v.CommissionRateBPS, v.CommissionAmount, v.CommissionCurrency, v.UpdatedAt}
}
func affiliatePageDTO(v admin.Page[admin.AffiliateLink], page, pageSize int) adminPageResponse[affiliateAdminResponse] {
	items := make([]affiliateAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, affiliateDTO(item))
	}
	return pageDTO(items, v.Total, page, pageSize)
}
func adminRecommendationDTO(v admin.Recommendation) recommendationAdminResponse {
	return recommendationAdminResponse{v.ID, v.SessionID, v.UserEmail, v.Goal, v.Experience, v.SessionStatus, v.ObjectiveScore, v.TotalPriceMinor, v.Currency, v.PolicyVersion, v.EngineVersion, v.CreatedAt}
}
func recommendationsPageDTO(v admin.Page[admin.Recommendation], page, pageSize int) adminPageResponse[recommendationAdminResponse] {
	items := make([]recommendationAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, adminRecommendationDTO(item))
	}
	return pageDTO(items, v.Total, page, pageSize)
}
func recommendationDetailDTO(v admin.RecommendationDetail) recommendationDetailAdminResponse {
	items := make([]recommendationItemAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		reasons := make([]recommendationReasonResponse, 0, len(item.Reasons))
		for _, reason := range item.Reasons {
			reasons = append(reasons, recommendationReasonResponse{reason.Code, reason.Message, reason.Dimension, reason.Score})
		}
		items = append(items, recommendationItemAdminResponse{item.ProductID, item.ProductName, item.ItemType, item.Rank, item.Quantity, item.ObjectiveScore, item.ReasonCode, item.ReasonSummary, item.RejectionCode, reasons})
	}
	s := v.Scores
	return recommendationDetailAdminResponse{adminRecommendationDTO(v.Recommendation), adminScoreBreakdownResponse{s.Goal, s.Budget, s.Space, s.Experience, s.Preference, s.Quality, s.Value, s.Durability, s.Compatibility, s.Portability, s.Noise}, items}
}
func usersPageDTO(v admin.Page[admin.User], page, pageSize int) adminPageResponse[userAdminResponse] {
	items := make([]userAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, userAdminResponse{item.ID, item.Email, item.Status, item.Roles, item.LastLoginAt, item.CreatedAt})
	}
	return pageDTO(items, v.Total, page, pageSize)
}
func eventsPageDTO(v admin.Page[admin.Event], page, pageSize int) adminPageResponse[eventAdminResponse] {
	items := make([]eventAdminResponse, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, eventAdminResponse{item.ID, item.Name, item.UserID, item.AnonymousID, item.SessionID, item.Surface, item.Properties, item.PagePath, item.TrafficSource, item.TrafficMedium, item.Campaign, item.ReferrerHost, item.ConsentState, item.OccurredAt})
	}
	return pageDTO(items, v.Total, page, pageSize)
}
func pageDTO[T any](items []T, total int64, page, pageSize int) adminPageResponse[T] {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return adminPageResponse[T]{items, page, pageSize, total, totalPages}
}
