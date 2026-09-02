package domain

import (
	"encoding/json"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

type Counts struct {
	Products        int64
	Published       int64
	Offers          int64
	ActiveOffers    int64
	Users           int64
	Recommendations int64
}

type AnalyticsCounts struct {
	RecommendationStarts     int64
	CompletedRecommendations int64
	ProductViews             int64
	AffiliateClicks          int64
	SavedProducts            int64
	SavedSetups              int64
}

type Dashboard struct {
	Counts    Counts
	Analytics AnalyticsCounts
	Readiness Readiness
}

// Readiness answers the question the counts do not: how much of the published
// catalog can actually earn. A published product with no active offer, or an
// active offer with no active affiliate link, is a page that costs traffic and
// returns nothing.
type Readiness struct {
	PublishedProducts    int64
	WithoutActiveOffer   int64
	WithoutAffiliateLink int64
	EarningReady         int64
	CommerceProviders    int64
	PublishedContent     int64
	Blocked              []BlockedProduct
}

type BlockedReason string

const (
	BlockedNoActiveOffer   BlockedReason = "no_active_offer"
	BlockedNoAffiliateLink BlockedReason = "no_affiliate_link"
)

type BlockedProduct struct {
	ID     catalog.ProductID
	Name   string
	Slug   string
	Reason BlockedReason
}

type ProductPage struct {
	Items []catalog.Product
	Total int64
}

type ProductInput struct {
	CategoryID  catalog.CategoryID
	BrandID     catalog.BrandID
	Name        string
	Slug        string
	Description string
	Price       catalog.Money
	Billing     catalog.Billing
	// IsPhysical mirrors the target category. A non-physical product leaves
	// the physical attributes below at their zero value and they are stored
	// as nulls.
	IsPhysical       bool
	Dimensions       catalog.Dimensions
	WeightGrams      int64
	MaxCapacityGrams *int64
	Material         string
	WarrantyMonths   int16
	Scores           catalog.Scores
}

type ImageInput struct {
	URL       string
	AltText   string
	SortOrder int
	IsPrimary bool
}

type ImageUpload struct {
	Data      []byte
	MIMEType  string
	AltText   string
	SortOrder int
	IsPrimary bool
}

type AttributeInput struct {
	Key          string
	Type         catalog.AttributeType
	NumericValue *float64
	TextValue    *string
	BooleanValue *bool
	Unit         *string
	IsFilterable bool
}

type Category struct {
	ID        string
	Name      string
	Slug      string
	IsActive  bool
	Products  int64
	UpdatedAt time.Time
}

type Brand struct {
	ID        string
	Name      string
	Slug      string
	IsActive  bool
	Products  int64
	UpdatedAt time.Time
}

type Merchant struct {
	ID          string
	Name        string
	Slug        string
	WebsiteURL  string
	CountryCode string
	TrustScore  int16
	Status      string
	Offers      int64
	UpdatedAt   time.Time
}

type Offer struct {
	ID              string
	MerchantID      string
	MerchantName    string
	ProductID       string
	ProductName     string
	MerchantSKU     string
	ProductURL      string
	PriceMinor      int64
	ShippingMinor   int64
	Currency        string
	Availability    string
	Condition       string
	IsActive        bool
	LastCheckedAt   time.Time
	ExpiresAt       *time.Time
	FreshnessStatus string
	AffiliateLinks  int64
	UpdatedAt       time.Time
}

type OfferInput struct {
	MerchantID    string
	ProductID     string
	MerchantSKU   string
	ProductURL    string
	PriceMinor    int64
	ShippingMinor int64
	Currency      string
	Availability  string
	Condition     string
	IsActive      bool
	Affiliate     *AffiliateLinkInput
}

type AffiliateLink struct {
	ID                 string
	OfferID            string
	ProductName        string
	MerchantName       string
	Provider           string
	DestinationURL     string
	ExternalReference  *string
	DisclosureLabel    string
	IsActive           bool
	Priority           int16
	ProgramID          *string
	CommissionType     string
	CommissionRateBPS  *int
	CommissionAmount   *int64
	CommissionCurrency *string
	UpdatedAt          time.Time
}

type AffiliateLinkInput struct {
	Provider           string
	DestinationURL     string
	ExternalReference  *string
	DisclosureLabel    string
	IsActive           bool
	Priority           int16
	ProgramID          *string
	CommissionType     string
	CommissionRateBPS  *int
	CommissionAmount   *int64
	CommissionCurrency *string
}

type Recommendation struct {
	ID              string
	SessionID       string
	UserEmail       *string
	Goal            string
	Experience      string
	SessionStatus   string
	ObjectiveScore  int16
	TotalPriceMinor int64
	Currency        string
	PolicyVersion   string
	EngineVersion   string
	CreatedAt       time.Time
}

type RecommendationItem struct {
	ProductID      string
	ProductName    string
	ItemType       string
	Rank           int16
	Quantity       int16
	ObjectiveScore int16
	ReasonCode     string
	ReasonSummary  string
	RejectionCode  *string
	Reasons        []RecommendationReason
}

type RecommendationReason struct {
	Code      string
	Message   string
	Dimension string
	Score     int16
}

type ScoreBreakdown struct {
	Goal          int16
	Budget        int16
	Space         int16
	Experience    int16
	Preference    int16
	Quality       int16
	Value         int16
	Durability    int16
	Compatibility int16
	Portability   int16
	Noise         int16
}

type RecommendationDetail struct {
	Recommendation Recommendation
	Scores         ScoreBreakdown
	Items          []RecommendationItem
}

type User struct {
	ID          string
	Email       string
	Status      string
	Roles       []string
	LastLoginAt *time.Time
	CreatedAt   time.Time
}

type Event struct {
	ID            string
	Name          string
	UserID        *string
	AnonymousID   *string
	SessionID     *string
	Surface       string
	Properties    json.RawMessage
	PagePath      *string
	TrafficSource *string
	TrafficMedium *string
	Campaign      *string
	ReferrerHost  *string
	ConsentState  string
	OccurredAt    time.Time
}

type References struct {
	Categories []Category
	Brands     []Brand
	Merchants  []Merchant
	Products   []ProductReference
}

type ProductReference struct {
	ID   string
	Name string
	Slug string
}

type Page[T any] struct {
	Items []T
	Total int64
}
