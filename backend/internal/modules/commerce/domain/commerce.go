package domain

import (
	"errors"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

type MerchantID string
type OfferID string
type AffiliateLinkID string
type PromotionSlug string

type Merchant struct {
	ID          MerchantID
	Name        string
	Slug        string
	WebsiteURL  string
	CountryCode string
	TrustScore  int16
	Status      string
}

type AffiliateLink struct {
	ID                AffiliateLinkID
	Provider          string
	DestinationURL    string
	ExternalReference *string
	DisclosureLabel   string
	IsActive          bool
	Priority          int16
	ProgramID         *string
	Commission        CommissionMetadata
}

type CommissionMetadata struct {
	Type        string
	RateBPS     *int
	AmountMinor *int64
	Currency    *string
}

type Offer struct {
	ID                 OfferID
	Merchant           Merchant
	ProductID          catalog.ProductID
	MerchantSKU        string
	ProductURL         string
	Price              catalog.Money
	ShippingMinor      int64
	Availability       string
	Condition          string
	LastCheckedAt      time.Time
	ProviderObservedAt *time.Time
	ImportedAt         *time.Time
	ExpiresAt          *time.Time
	IsActive           bool
	AffiliateLinks     []AffiliateLink
}

// PurchasableOffer is the least an interface needs in order to draw a working
// vendor button: which offer to redirect through, whose name goes on it, and
// the disclosure that must sit beside it.
//
// It exists so a grid of products can be answered in one query. Asking for the
// full Offer per card is a request per card, and the catalog listing draws
// twenty-four of them.
type PurchasableOffer struct {
	OfferID          OfferID
	MerchantName     string
	DisclosureLabel  string
	Price            catalog.Money
	LandedPriceMinor int64
	// LastCheckedAt is when the price was last read from the vendor's page.
	// A surface that prints the price owes the reader its date; the live
	// offers listing prints fifteen of them and asks in one query.
	LastCheckedAt time.Time
}

type AffiliateClick struct {
	OfferID          OfferID
	LinkID           AffiliateLinkID
	PromotionSlug    PromotionSlug
	UserID           *string
	AnonymousID      *string
	SessionID        *string
	Source           string
	Campaign         *string
	Referrer         *string
	TrafficSource    *string
	TrafficMedium    *string
	ReferrerHost     *string
	RecommendationID *string
	RequestID        *string
	IdempotencyKey   *string
	Classification   ClickClassification
	IsCountable      bool
	UserAgentHash    *string
	RetentionExpires time.Time
}

type ClickClassification string

const (
	ClickHuman    ClickClassification = "human"
	ClickBot      ClickClassification = "bot"
	ClickPrefetch ClickClassification = "prefetch"
	ClickUnknown  ClickClassification = "unknown"
)

type ResolvedAffiliateDestination struct {
	AffiliateLinkID    AffiliateLinkID
	OfferID            OfferID
	ProductID          catalog.ProductID
	RecommendationItem *string
	DestinationURL     string
}

// ResolvedPromotionDestination carries only what the standalone promotion
// redirect needs. It deliberately has no catalog product or recommendation
// fields: promotional economics cannot become recommendation inputs.
type ResolvedPromotionDestination struct {
	PromotionID    string
	PromotionSlug  PromotionSlug
	DestinationURL string
}

type AffiliateRedirectResult struct {
	DestinationURL string
	TrackingError  error
}

func (offer Offer) Validate() error {
	if offer.ID == "" || offer.ProductID == "" || offer.Merchant.ID == "" {
		return errors.New("offer identifiers are required")
	}
	if offer.Price.AmountMinor < 0 || offer.ShippingMinor < 0 {
		return errors.New("offer price and shipping cannot be negative")
	}
	if len(offer.Price.Currency) != 3 || offer.Price.Currency != strings.ToUpper(offer.Price.Currency) {
		return errors.New("offer currency is invalid")
	}
	return nil
}

func (offer Offer) LandedPriceMinor() int64 {
	return offer.Price.AmountMinor + offer.ShippingMinor
}
