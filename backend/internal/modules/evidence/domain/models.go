package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type SourceType string

const (
	SourceManufacturer SourceType = "manufacturer_documentation"
	SourceIndependent  SourceType = "independent_testing"
	SourceMerchant     SourceType = "verified_merchant_data"
	SourceEditorial    SourceType = "editorial_assessment"
	SourceDemo         SourceType = "demo_fixture"
)

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewVerified ReviewStatus = "verified"
	ReviewRejected ReviewStatus = "rejected"
)

type WorkflowStatus string

const (
	WorkflowDraft      WorkflowStatus = "draft"
	WorkflowInReview   WorkflowStatus = "in_review"
	WorkflowApproved   WorkflowStatus = "approved"
	WorkflowPublished  WorkflowStatus = "published"
	WorkflowRejected   WorkflowStatus = "rejected"
	WorkflowSuperseded WorkflowStatus = "superseded"
)

type Classification string

const (
	ClassificationVerified     Classification = "verified_fact"
	ClassificationManufacturer Classification = "manufacturer_claim"
	ClassificationMerchant     Classification = "merchant_observation"
	ClassificationEditorial    Classification = "editorial_assessment"
)

type Source struct {
	ID              string
	Type            SourceType
	Title           string
	Publisher       string
	URL             *string
	IsFictional     bool
	ReviewStatus    ReviewStatus
	ReviewerUserID  *identity.UserID
	ReviewedAt      *time.Time
	ReviewNote      string
	CreatedByUserID *identity.UserID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type SourceInput struct {
	Type        SourceType
	Title       string
	Publisher   string
	URL         *string
	IsFictional bool
}

func (input SourceInput) Validate() error {
	if input.Type != SourceManufacturer && input.Type != SourceIndependent &&
		input.Type != SourceMerchant && input.Type != SourceEditorial && input.Type != SourceDemo {
		return errors.New("unsupported evidence source type")
	}
	if len(strings.TrimSpace(input.Title)) < 1 || len(strings.TrimSpace(input.Title)) > 240 ||
		len(strings.TrimSpace(input.Publisher)) < 1 || len(strings.TrimSpace(input.Publisher)) > 180 {
		return errors.New("source title and publisher are required")
	}
	if input.IsFictional != (input.Type == SourceDemo) {
		return errors.New("only demo fixtures may be fictional and demo fixtures must be fictional")
	}
	if input.URL == nil {
		if input.Type != SourceDemo {
			return errors.New("non-demo source URL is required")
		}
		return nil
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(*input.URL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("source URL must be an absolute HTTPS URL")
	}
	return nil
}

type Observation struct {
	ID         string
	SourceID   string
	ProductID  catalog.ProductID
	ObservedAt time.Time
	ExpiresAt  *time.Time
	Confidence int16
	Notes      string
	CreatedAt  time.Time
}

type ObservationInput struct {
	SourceID   string
	ProductID  catalog.ProductID
	ObservedAt time.Time
	ExpiresAt  *time.Time
	Confidence int16
	Notes      string
}

func (input ObservationInput) Validate(now time.Time) error {
	if input.SourceID == "" || input.ProductID == "" || input.ObservedAt.IsZero() ||
		input.ObservedAt.After(now.Add(5*time.Minute)) || input.Confidence < 0 || input.Confidence > 100 ||
		len(input.Notes) > 4000 {
		return errors.New("invalid evidence observation")
	}
	if input.ExpiresAt != nil && !input.ExpiresAt.After(input.ObservedAt) {
		return errors.New("evidence expiration must follow its observation")
	}
	return nil
}

type FactLink struct {
	FactKey        string
	ObservationID  string
	Classification Classification
}

type ScoreRationale struct {
	ScoreKey      string
	Rationale     string
	ObservationID string
}

type RevisionInput struct {
	Product    catalog.Product
	FactLinks  []FactLink
	Scores     catalog.Scores
	Rationales []ScoreRationale
}

type Revision struct {
	FactRevisionID    string
	ScoreRevisionID   string
	ProductID         catalog.ProductID
	FactVersion       int
	ScoreVersion      int
	Status            WorkflowStatus
	CreatedByUserID   *identity.UserID
	SubmittedByUserID *identity.UserID
	ReviewedByUserID  *identity.UserID
	PublishedByUserID *identity.UserID
	CreatedAt         time.Time
	SubmittedAt       *time.Time
	ReviewedAt        *time.Time
	PublishedAt       *time.Time
	ValidUntil        *time.Time
	ReviewNote        string
}

type Provenance struct {
	FactKey        string
	ScoreKey       string
	Classification Classification
	Rationale      string
	Observation    Observation
	Source         Source
}

type AuditEvent struct {
	Action     string
	ActorEmail *string
	Changes    map[string]string
	OccurredAt time.Time
}

type ProductGovernance struct {
	ProductID                catalog.ProductID
	ProductName              string
	Status                   catalog.ProductStatus
	PublishedFactRevisionID  *string
	PublishedScoreRevisionID *string
	Revisions                []Revision
	Provenance               []Provenance
	Audit                    []AuditEvent
}

var factKeys = map[string]bool{
	"category": true, "brand": true, "name": true, "slug": true,
	"description": true, "price": true, "dimensions": true, "weight": true,
	"max_capacity": true, "material": true, "warranty": true,
}

var scoreKeys = map[string]bool{
	"quality": true, "value": true, "durability": true, "beginner": true,
	"advanced": true, "apartment": true, "noise": true, "portability": true,
}

func (input RevisionInput) Validate() error {
	if err := input.Product.Validate(); err != nil {
		return err
	}
	if err := input.Scores.Validate(); err != nil {
		return err
	}
	requiredFacts := map[string]bool{
		"category": true, "brand": true, "name": true, "slug": true,
		"description": true, "price": true, "warranty": true,
	}
	// Provenance is required for populated facts. A non-physical product has
	// no dimensions, weight or material to attest to, so demanding provenance
	// for them would mean citing a source for a null.
	if input.Product.IsPhysical {
		requiredFacts["dimensions"] = true
		requiredFacts["weight"] = true
		requiredFacts["material"] = true
	}
	if input.Product.MaxCapacityGrams != nil {
		requiredFacts["max_capacity"] = true
	}
	seenFacts := make(map[string]bool, len(input.FactLinks))
	for _, link := range input.FactLinks {
		if !factKeys[link.FactKey] || link.ObservationID == "" ||
			(link.Classification != ClassificationVerified && link.Classification != ClassificationManufacturer &&
				link.Classification != ClassificationMerchant && link.Classification != ClassificationEditorial) {
			return errors.New("invalid fact provenance")
		}
		seenFacts[link.FactKey] = true
	}
	for key := range requiredFacts {
		if !seenFacts[key] {
			return errors.New("every populated recommendation-critical fact requires provenance")
		}
	}
	seenScores := make(map[string]bool, len(input.Rationales))
	for _, rationale := range input.Rationales {
		if !scoreKeys[rationale.ScoreKey] || rationale.ObservationID == "" ||
			len(strings.TrimSpace(rationale.Rationale)) < 1 || len(rationale.Rationale) > 2000 {
			return errors.New("invalid score rationale")
		}
		seenScores[rationale.ScoreKey] = true
	}
	for key := range scoreKeys {
		if !seenScores[key] {
			return errors.New("every score requires an evidence-backed rationale")
		}
	}
	return nil
}
