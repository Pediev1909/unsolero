package domain

import (
	"errors"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
)

var ErrUnsupportedCategory = errors.New("category is not supported by the active recommendation policy")
var ErrProductPolicyMissing = errors.New("product has no active recommendation policy")

type PolicyWorkflowStatus string

const (
	PolicyDraft    PolicyWorkflowStatus = "draft"
	PolicyInReview PolicyWorkflowStatus = "in_review"
	PolicyApproved PolicyWorkflowStatus = "approved"
	PolicyActive   PolicyWorkflowStatus = "active"
	PolicyRetired  PolicyWorkflowStatus = "retired"
	PolicyRejected PolicyWorkflowStatus = "rejected"
)

type PolicySummary struct {
	Version       string
	VerticalKey   string
	Status        PolicyWorkflowStatus
	CategoryCount int
	ProductCount  int
	CreatedAt     time.Time
	ActivatedAt   *time.Time
	ReviewNote    string
}

type CategoryPolicy struct {
	CategorySlug     string
	Supported        bool
	Capabilities     []Capability
	RedundancyGroups []string
}

type ProductPolicy struct {
	ProductID        catalog.ProductID
	FactRevisionID   string
	ScoreRevisionID  string
	Capabilities     []Capability
	Requires         []Capability
	CompatibleWith   []Capability
	IncompatibleWith []Capability
	GoalSupport      []GoalSupport
	PreferenceTags   []TrainingPreference
	RedundancyGroups []string
	Space            SpaceProfile
}

type Policy struct {
	Config     Config
	Categories map[string]CategoryPolicy
	Products   map[catalog.ProductID]ProductPolicy
}

func (policy Policy) Candidate(product catalog.Product) (CandidateSnapshot, error) {
	category, exists := policy.Categories[product.CategorySlug]
	if !exists || !category.Supported {
		return CandidateSnapshot{}, ErrUnsupportedCategory
	}
	configured, exists := policy.Products[product.ID]
	if !exists || configured.FactRevisionID != product.FactRevisionID ||
		configured.ScoreRevisionID != product.ScoreRevisionID {
		return CandidateSnapshot{}, ErrProductPolicyMissing
	}
	candidate := CandidateFromProduct(product)
	candidate.PolicyVersion = policy.Config.PolicyVersion
	candidate.Capabilities = mergedCapabilities(category.Capabilities, configured.Capabilities)
	candidate.Requires = mergedCapabilities(configured.Requires)
	candidate.CompatibleWith = mergedCapabilities(configured.CompatibleWith)
	candidate.IncompatibleWith = mergedCapabilities(configured.IncompatibleWith)
	candidate.GoalSupport = append([]GoalSupport(nil), configured.GoalSupport...)
	candidate.PreferenceTags = append([]TrainingPreference(nil), configured.PreferenceTags...)
	candidate.RedundancyGroups = mergedStrings(category.RedundancyGroups, configured.RedundancyGroups)
	candidate.Space = configured.Space
	return candidate, nil
}

func (policy Policy) EnrichInput(input Input) Input {
	result := input
	result.ExistingEquipment = append([]ExistingEquipment(nil), input.ExistingEquipment...)
	for index := range result.ExistingEquipment {
		category, exists := policy.Categories[result.ExistingEquipment[index].CategorySlug]
		if !exists || !category.Supported {
			continue
		}
		result.ExistingEquipment[index].Capabilities = mergedCapabilities(
			result.ExistingEquipment[index].Capabilities, category.Capabilities)
		result.ExistingEquipment[index].RedundancyGroups = mergedStrings(
			result.ExistingEquipment[index].RedundancyGroups, category.RedundancyGroups)
	}
	return result
}
