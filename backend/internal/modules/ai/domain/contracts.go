package domain

import (
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/domain"
)

var (
	ErrInvalidInput          = errors.New("invalid AI input")
	ErrInvalidProviderOutput = errors.New("invalid AI provider output")
)

type RequirementField string

const (
	FieldGoal                RequirementField = "goal"
	FieldExperience          RequirementField = "experience"
	FieldBudget              RequirementField = "budget"
	FieldAvailableSpace      RequirementField = "available_space"
	FieldExistingEquipment   RequirementField = "existing_equipment"
	FieldTrainingPreferences RequirementField = "training_preferences"
	FieldPriorities          RequirementField = "priorities"
	FieldFreeText            RequirementField = "free_text"
)

type AmbiguityKind string

const (
	AmbiguityMissing     AmbiguityKind = "missing"
	AmbiguityConflicting AmbiguityKind = "conflicting"
	AmbiguityUnclear     AmbiguityKind = "unclear"
)

type UnderstandUserInputRequest struct {
	Text   string `json:"text"`
	Locale string `json:"locale"`
}

type Ambiguity struct {
	Field RequirementField `json:"field"`
	Kind  AmbiguityKind    `json:"kind"`
}

type UserInputUnderstanding struct {
	Language           string             `json:"language"`
	Confidence         int16              `json:"confidence"`
	RecognizedFields   []RequirementField `json:"recognized_fields"`
	Ambiguities        []Ambiguity        `json:"ambiguities"`
	NeedsClarification bool               `json:"needs_clarification"`
}

// RequirementsDraft is intentionally incomplete. Provider output is confirmed
// or completed before it is converted to recommendation.Input and ranked.
type RequirementsDraft struct {
	Goal                *planning.Goal                      `json:"goal,omitempty"`
	Experience          *planning.ExperienceLevel           `json:"experience,omitempty"`
	Budget              *catalog.Money                      `json:"budget,omitempty"`
	AvailableSpace      *recommendation.AvailableSpace      `json:"available_space,omitempty"`
	ExistingEquipment   []recommendation.ExistingEquipment  `json:"existing_equipment"`
	TrainingPreferences []recommendation.TrainingPreference `json:"training_preferences"`
	Priorities          []recommendation.Priority           `json:"priorities"`
	FreeText            string                              `json:"free_text"`
}

type ExtractRequirementsRequest struct {
	Text   string            `json:"text"`
	Locale string            `json:"locale"`
	Known  RequirementsDraft `json:"known"`
}

type ClarifyingQuestionRequest struct {
	Requirements RequirementsDraft `json:"requirements"`
	Ambiguities  []Ambiguity       `json:"ambiguities"`
	Locale       string            `json:"locale"`
}

type ClarifyingPrompt string

const (
	PromptPrimaryGoal         ClarifyingPrompt = "ask_primary_goal"
	PromptExperience          ClarifyingPrompt = "ask_experience"
	PromptBudget              ClarifyingPrompt = "ask_budget"
	PromptAvailableSpace      ClarifyingPrompt = "ask_available_space"
	PromptExistingEquipment   ClarifyingPrompt = "ask_existing_equipment"
	PromptTrainingPreferences ClarifyingPrompt = "ask_training_preferences"
	PromptPriorities          ClarifyingPrompt = "ask_priorities"
	PromptMoreDetail          ClarifyingPrompt = "ask_more_detail"
)

// ClarifyingQuestion selects trusted localized copy. It deliberately contains
// no model-authored prose or product fields.
type ClarifyingQuestion struct {
	Field  RequirementField `json:"field"`
	Prompt ClarifyingPrompt `json:"prompt"`
}

type FactKey string

const (
	FactPrice       FactKey = "price"
	FactDimensions  FactKey = "dimensions"
	FactWeight      FactKey = "weight"
	FactMaxCapacity FactKey = "max_capacity"
	FactMaterial    FactKey = "material"
	FactWarranty    FactKey = "warranty"
	FactQuality     FactKey = "quality_score"
	FactValue       FactKey = "value_score"
	FactDurability  FactKey = "durability_score"
	FactBeginner    FactKey = "beginner_score"
	FactAdvanced    FactKey = "advanced_score"
	FactApartment   FactKey = "apartment_score"
	FactNoise       FactKey = "noise_score"
	FactPortability FactKey = "portability_score"
)

// ProductSnapshot is an immutable, application-supplied view of canonical
// catalog facts. Providers can inspect it but cannot return replacements.
type ProductSnapshot struct {
	ID               catalog.ProductID  `json:"id"`
	Name             string             `json:"name"`
	Category         string             `json:"category"`
	Brand            string             `json:"brand"`
	Price            catalog.Money      `json:"price"`
	Dimensions       catalog.Dimensions `json:"dimensions"`
	WeightGrams      int64              `json:"weight_grams"`
	MaxCapacityGrams *int64             `json:"max_capacity_grams,omitempty"`
	Material         string             `json:"material"`
	WarrantyMonths   int16              `json:"warranty_months"`
	Scores           catalog.Scores     `json:"scores"`
}

type DeterministicProduct struct {
	Product        ProductSnapshot `json:"product"`
	Rank           int             `json:"rank"`
	ObjectiveScore int             `json:"objective_score"`
	ReasonCodes    []string        `json:"reason_codes"`
}

type RejectedProduct struct {
	Product ProductSnapshot `json:"product"`
	Code    string          `json:"code"`
}

// RecommendationSnapshot can only be constructed after the deterministic
// engine has made eligibility, ranking, and price decisions.
type RecommendationSnapshot struct {
	Requirements   RequirementsDraft      `json:"requirements"`
	Selected       []DeterministicProduct `json:"selected"`
	Alternatives   []DeterministicProduct `json:"alternatives"`
	Rejected       []RejectedProduct      `json:"rejected"`
	TotalCost      catalog.Money          `json:"total_cost"`
	ObjectiveScore int                    `json:"objective_score"`
	PolicyVersion  string                 `json:"policy_version"`
	EngineVersion  string                 `json:"engine_version"`
}

type ExplanationStyle string

const (
	StyleBalanced    ExplanationStyle = "balanced"
	StyleBudget      ExplanationStyle = "budget_focused"
	StyleSpace       ExplanationStyle = "space_focused"
	StylePerformance ExplanationStyle = "performance_focused"
)

type ExplainRecommendationRequest struct {
	Recommendation RecommendationSnapshot `json:"recommendation"`
	Locale         string                 `json:"locale"`
}

// ExplanationPlan carries references, not prose claims. The application
// renders these references using canonical facts and deterministic reasons.
type ExplanationPlan struct {
	Style  ExplanationStyle         `json:"style"`
	Points []ProductExplanationPlan `json:"points"`
}

type ProductExplanationPlan struct {
	ProductID   catalog.ProductID `json:"product_id"`
	ReasonCodes []string          `json:"reason_codes"`
	FactKeys    []FactKey         `json:"fact_keys"`
}

type RefineRecommendationRequest struct {
	Instruction    string                 `json:"instruction"`
	Current        RequirementsDraft      `json:"current"`
	Recommendation RecommendationSnapshot `json:"recommendation"`
	Locale         string                 `json:"locale"`
}

// Refinement changes requirements only. Callers must rerun the deterministic
// engine; a provider cannot directly add, remove, or reorder products.
type Refinement struct {
	Requirements  RequirementsDraft  `json:"requirements"`
	ChangedFields []RequirementField `json:"changed_fields"`
}

type CompareProductsRequest struct {
	Products []ProductSnapshot `json:"products"`
	Focus    []FactKey         `json:"focus"`
	Locale   string            `json:"locale"`
}

type HighlightKind string

const (
	HighlightStrength HighlightKind = "strength"
	HighlightTradeoff HighlightKind = "tradeoff"
)

type ComparisonHighlight struct {
	ProductID catalog.ProductID `json:"product_id"`
	FactKey   FactKey           `json:"fact_key"`
	Kind      HighlightKind     `json:"kind"`
}

// ComparisonPlan selects trusted facts to display. It contains no model-made
// values, product descriptions, prices, or specifications.
type ComparisonPlan struct {
	OrderedProductIDs []catalog.ProductID   `json:"ordered_product_ids"`
	FactKeys          []FactKey             `json:"fact_keys"`
	Highlights        []ComparisonHighlight `json:"highlights"`
}
