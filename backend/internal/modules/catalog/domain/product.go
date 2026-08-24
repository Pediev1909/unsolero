package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	slugPattern         = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	attributeKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

type ProductID string
type CategoryID string
type BrandID string

type ProductStatus string

const (
	ProductStatusDraft        ProductStatus = "draft"
	ProductStatusPublished    ProductStatus = "published"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

type Money struct {
	AmountMinor int64
	Currency    string
}

func (money Money) Validate() error {
	if money.AmountMinor < 0 {
		return errors.New("money amount cannot be negative")
	}
	if len(money.Currency) != 3 || money.Currency != strings.ToUpper(money.Currency) {
		return errors.New("currency must be a three-letter uppercase code")
	}
	return nil
}

type Dimensions struct {
	LengthMM int64
	WidthMM  int64
	HeightMM int64
}

func (dimensions Dimensions) Validate() error {
	if dimensions.LengthMM <= 0 || dimensions.WidthMM <= 0 || dimensions.HeightMM <= 0 {
		return errors.New("all product dimensions must be positive")
	}
	return nil
}

type Scores struct {
	Quality     int16
	Value       int16
	Durability  int16
	Beginner    int16
	Advanced    int16
	Apartment   int16
	Noise       int16
	Portability int16
}

func (scores Scores) Validate() error {
	values := []struct {
		name  string
		value int16
	}{
		{name: "quality", value: scores.Quality},
		{name: "value", value: scores.Value},
		{name: "durability", value: scores.Durability},
		{name: "beginner", value: scores.Beginner},
		{name: "advanced", value: scores.Advanced},
		{name: "apartment", value: scores.Apartment},
		{name: "noise", value: scores.Noise},
		{name: "portability", value: scores.Portability},
	}
	for _, score := range values {
		if score.value < 0 || score.value > 100 {
			return fmt.Errorf("%s score must be between 0 and 100", score.name)
		}
	}
	return nil
}

type AttributeType string

const (
	AttributeTypeNumber  AttributeType = "number"
	AttributeTypeText    AttributeType = "text"
	AttributeTypeBoolean AttributeType = "boolean"
)

type Attribute struct {
	ID           string
	Key          string
	Type         AttributeType
	NumericValue *float64
	TextValue    *string
	BooleanValue *bool
	Unit         *string
	IsFilterable bool
}

func (attribute Attribute) Validate() error {
	if !attributeKeyPattern.MatchString(attribute.Key) {
		return errors.New("attribute key must be lower snake case")
	}

	switch attribute.Type {
	case AttributeTypeNumber:
		if attribute.NumericValue == nil || attribute.TextValue != nil || attribute.BooleanValue != nil {
			return errors.New("number attribute must contain only a numeric value")
		}
	case AttributeTypeText:
		if attribute.TextValue == nil || attribute.NumericValue != nil || attribute.BooleanValue != nil || attribute.Unit != nil {
			return errors.New("text attribute must contain only a text value")
		}
	case AttributeTypeBoolean:
		if attribute.BooleanValue == nil || attribute.NumericValue != nil || attribute.TextValue != nil || attribute.Unit != nil {
			return errors.New("boolean attribute must contain only a boolean value")
		}
	default:
		return errors.New("unsupported attribute type")
	}

	return nil
}

type Product struct {
	ID           ProductID
	CategoryID   CategoryID
	CategoryName string
	CategorySlug string
	BrandID      BrandID
	BrandName    string
	BrandSlug    string
	Name         string
	Slug         string
	Description  string
	Price        Money
	// IsPhysical mirrors the product category. A non-physical product (for
	// example a software subscription) carries no dimensions, weight or
	// material, and those fields stay at their zero value.
	IsPhysical       bool
	Dimensions       Dimensions
	WeightGrams      int64
	MaxCapacityGrams *int64
	Material         string
	WarrantyMonths   int16
	Scores           Scores
	Status           ProductStatus
	FactRevisionID   string
	ScoreRevisionID  string
	Evidence         []ProductEvidence
	Images           []ProductImage
	Attributes       []Attribute
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProductEvidence is a public, non-commercial description of why a fact or
// assessment may be trusted. Affiliate and merchant payout data cannot be
// represented here by design.
type ProductEvidence struct {
	FactKey        string
	Classification string
	SourceType     string
	SourceTitle    string
	SourceURL      *string
	ObservedAt     time.Time
	ExpiresAt      *time.Time
	Confidence     int16
	IsFictional    bool
}

type SuitabilityInsight struct {
	Key   string
	Label string
	Score int16
}

// Suitability omits the dimensions that only mean something for a physical
// product. Reporting "Apartment: 0" for a subscription reads as a damning score
// rather than as an inapplicable one, and portability means keeping your data
// rather than carrying the thing.
func (product Product) Suitability() []SuitabilityInsight {
	if !product.IsPhysical {
		return []SuitabilityInsight{
			{Key: "beginner", Label: "Easy to adopt", Score: product.Scores.Beginner},
			{Key: "advanced", Label: "Depth for power users", Score: product.Scores.Advanced},
			{Key: "portable", Label: "Data portability", Score: product.Scores.Portability},
		}
	}
	return []SuitabilityInsight{
		{Key: "beginner", Label: "Beginner", Score: product.Scores.Beginner},
		{Key: "advanced", Label: "Advanced", Score: product.Scores.Advanced},
		{Key: "apartment", Label: "Apartment", Score: product.Scores.Apartment},
		{Key: "quiet", Label: "Quiet use", Score: product.Scores.Noise},
		{Key: "portable", Label: "Portable", Score: product.Scores.Portability},
	}
}

func (product Product) Strengths() []SuitabilityInsight {
	return filterInsights(product.scoreInsights(), func(score int16) bool { return score >= 85 })
}

func (product Product) Considerations() []SuitabilityInsight {
	return filterInsights(product.scoreInsights(), func(score int16) bool { return score <= 60 })
}

func (product Product) UseCases() []SuitabilityInsight {
	return filterInsights(product.Suitability(), func(score int16) bool { return score >= 85 })
}

// scoreInsights feeds Strengths and Considerations. Considerations reports any
// dimension at or below 60, so leaving the inapplicable physical dimensions in
// would make every software product look like it had three serious weaknesses.
func (product Product) scoreInsights() []SuitabilityInsight {
	// The dimensions are shared; what they are called is not. A subscription
	// listed under "Build quality" and "Durability" reads as a page written
	// about something else, which is what these labels were written about.
	if !product.IsPhysical {
		return []SuitabilityInsight{
			{Key: "quality", Label: "Product quality", Score: product.Scores.Quality},
			{Key: "value", Label: "Value for money", Score: product.Scores.Value},
			{Key: "durability", Label: "Vendor stability", Score: product.Scores.Durability},
			{Key: "beginner", Label: "Easy to adopt", Score: product.Scores.Beginner},
			{Key: "advanced", Label: "Depth for power users", Score: product.Scores.Advanced},
			{Key: "portable", Label: "Data portability", Score: product.Scores.Portability},
		}
	}
	return []SuitabilityInsight{
		{Key: "quality", Label: "Build quality", Score: product.Scores.Quality},
		{Key: "value", Label: "Value", Score: product.Scores.Value},
		{Key: "durability", Label: "Durability", Score: product.Scores.Durability},
		{Key: "beginner", Label: "Beginner suitability", Score: product.Scores.Beginner},
		{Key: "advanced", Label: "Advanced suitability", Score: product.Scores.Advanced},
		{Key: "apartment", Label: "Apartment suitability", Score: product.Scores.Apartment},
		{Key: "quiet", Label: "Quiet operation", Score: product.Scores.Noise},
		{Key: "portable", Label: "Portability", Score: product.Scores.Portability},
	}
}

func filterInsights(
	insights []SuitabilityInsight,
	keep func(int16) bool,
) []SuitabilityInsight {
	filtered := make([]SuitabilityInsight, 0, len(insights))
	for _, insight := range insights {
		if keep(insight.Score) {
			filtered = append(filtered, insight)
		}
	}
	return filtered
}

type ProductImage struct {
	ID        string
	URL       string
	AltText   string
	SortOrder int
	IsPrimary bool
	WidthPX   *int
	HeightPX  *int
}

func (product Product) Validate() error {
	if product.ID == "" || product.CategoryID == "" || product.BrandID == "" {
		return errors.New("product and ownership identifiers are required")
	}
	if strings.TrimSpace(product.Name) == "" || strings.TrimSpace(product.Description) == "" {
		return errors.New("product name and description are required")
	}
	if !slugPattern.MatchString(product.Slug) {
		return errors.New("product slug is invalid")
	}
	if err := product.Price.Validate(); err != nil {
		return fmt.Errorf("price: %w", err)
	}
	if product.IsPhysical {
		if err := product.Dimensions.Validate(); err != nil {
			return fmt.Errorf("dimensions: %w", err)
		}
		if product.WeightGrams <= 0 {
			return errors.New("product weight must be positive")
		}
		if product.MaxCapacityGrams != nil && *product.MaxCapacityGrams <= 0 {
			return errors.New("maximum capacity must be positive when supplied")
		}
		if strings.TrimSpace(product.Material) == "" {
			return errors.New("material is required")
		}
	} else if product.Dimensions != (Dimensions{}) || product.WeightGrams != 0 ||
		product.MaxCapacityGrams != nil || strings.TrimSpace(product.Material) != "" {
		// Rejected rather than ignored: a stray footprint on a software
		// product would be scored as a real measurement by the engine.
		return errors.New("a non-physical product must not carry physical attributes")
	}
	if product.WarrantyMonths < 0 {
		return errors.New("warranty cannot be negative")
	}
	if err := product.Scores.Validate(); err != nil {
		return err
	}
	if product.Status != ProductStatusDraft && product.Status != ProductStatusPublished && product.Status != ProductStatusDiscontinued {
		return errors.New("product status is invalid")
	}
	for _, attribute := range product.Attributes {
		if err := attribute.Validate(); err != nil {
			return fmt.Errorf("attribute %q: %w", attribute.Key, err)
		}
	}
	return nil
}

type Category struct {
	ID          CategoryID
	ParentID    *CategoryID
	Name        string
	Slug        string
	Description string
	SortOrder   int
	IsActive    bool
	// A category exists before anything is filed under it. Until something is,
	// its page is a promise rather than a page, and neither the sitemap nor a
	// search engine should be told otherwise.
	PublishedProducts int
}

type Brand struct {
	ID          BrandID
	Name        string
	Slug        string
	Description string
	// See Category.PublishedProducts.
	PublishedProducts int
	WebsiteURL        *string
	CountryCode       *string
	IsActive          bool
}
