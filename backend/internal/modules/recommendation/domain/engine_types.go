package domain

import (
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

const (
	EngineVersion = "deterministic-v2"
	// MaximumCandidates is the explicit upper bound for a deterministic run. The
	// application must either load the complete catalog within this bound or fail;
	// silently scoring an arbitrary prefix would make recommendations untrustworthy.
	MaximumCandidates = 1000
	maxInputItems     = MaximumCandidates
	maxFreeText       = 1000
	maxMoneyMinor     = 1_000_000_000_000
	maxDimensionMM    = 100_000
)

var (
	ErrInvalidInput     = errors.New("invalid recommendation input")
	ErrInvalidCandidate = errors.New("invalid recommendation candidate")
	ErrInvalidConfig    = errors.New("invalid recommendation configuration")
)

type RecommendationEngine interface {
	Recommend(Input, []CandidateSnapshot) (Result, error)
}

type Input struct {
	Goal                planning.Goal
	Experience          planning.ExperienceLevel
	Budget              catalog.Money
	AvailableSpace      AvailableSpace
	ExistingEquipment   []ExistingEquipment
	TrainingPreferences []TrainingPreference
	Priorities          []Priority
	FreeText            string
}

type AvailableSpace struct {
	LengthMM        int64
	WidthMM         int64
	HeightMM        int64
	AccessWidthMM   *int64
	ApartmentLiving bool
}

type ExistingEquipment struct {
	Name             string
	CategorySlug     string
	Capabilities     []Capability
	RedundancyGroups []string
}

type TrainingPreference string

const (
	PreferenceDumbbells      TrainingPreference = "dumbbells"
	PreferenceBarbell        TrainingPreference = "barbell"
	PreferenceKettlebell     TrainingPreference = "kettlebell"
	PreferenceResistanceBand TrainingPreference = "resistance_bands"
	PreferenceBodyweight     TrainingPreference = "bodyweight"
	PreferenceCardio         TrainingPreference = "cardio"
	PreferenceLowImpact      TrainingPreference = "low_impact"
)

type Priority string

const (
	PriorityBudget      Priority = "budget"
	PriorityCompact     Priority = "compact"
	PriorityQuality     Priority = "quality"
	PriorityDurability  Priority = "durability"
	PriorityQuiet       Priority = "quiet"
	PriorityPortability Priority = "portability"
)

type Capability string

const (
	CapabilityResistanceTraining Capability = "resistance_training"
	CapabilityStrengthTraining   Capability = "strength_training"
	CapabilityHypertrophy        Capability = "hypertrophy"
	CapabilitySupportedTraining  Capability = "supported_training"
	CapabilityBarbellTraining    Capability = "barbell_training"
	CapabilityWeightPlates       Capability = "weight_plates"
	CapabilitySafeBarbell        Capability = "safe_barbell_training"
	CapabilityConditioning       Capability = "conditioning"
	CapabilityMobility           Capability = "mobility"
	CapabilityLowImpact          Capability = "low_impact"
	CapabilityPullUp             Capability = "pull_up"
	CapabilityAnchorPoint        Capability = "anchor_point"
)

type CandidateSnapshot struct {
	ProductID        catalog.ProductID
	FactRevisionID   string
	ScoreRevisionID  string
	Name             string
	CategorySlug     string
	PolicyVersion    string
	Price            catalog.Money
	Dimensions       catalog.Dimensions
	Scores           catalog.Scores
	GoalSupport      []GoalSupport
	PreferenceTags   []TrainingPreference
	RedundancyGroups []string
	Space            SpaceProfile
	Capabilities     []Capability
	Requires         []Capability
	CompatibleWith   []Capability
	IncompatibleWith []Capability
}

type GoalSupport struct {
	Goal  planning.Goal
	Score int
}

type Clearance struct {
	FrontMM int64
	BackMM  int64
	LeftMM  int64
	RightMM int64
	TopMM   int64
}

type SpatialEnvelope struct {
	LengthMM int64
	WidthMM  int64
	HeightMM int64
}

// SpaceProfile distinguishes known measurements from unknown measurements.
// Nil values are never interpreted as zero. OverlapGroup permits co-location
// only when the active policy explicitly assigns the same non-empty group.
type SpaceProfile struct {
	Footprint                  SpatialEnvelope
	StorageFootprint           *SpatialEnvelope
	OperatingClearance         *Clearance
	SafetyClearance            *Clearance
	MinimumRoomHeightMM        *int64
	MinimumAccessWidthMM       *int64
	OverlapGroup               string
	RequiresStorageFootprint   bool
	RequiresOperatingClearance bool
	RequiresSafetyClearance    bool
	RequiresAccessWidth        bool
}

func CandidateFromProduct(product catalog.Product) CandidateSnapshot {
	return CandidateSnapshot{
		ProductID: product.ID, FactRevisionID: product.FactRevisionID,
		ScoreRevisionID: product.ScoreRevisionID, Name: product.Name, CategorySlug: product.CategorySlug,
		Price: product.Price, Dimensions: product.Dimensions, Scores: product.Scores,
		Space: SpaceProfile{Footprint: SpatialEnvelope{LengthMM: product.Dimensions.LengthMM,
			WidthMM: product.Dimensions.WidthMM, HeightMM: product.Dimensions.HeightMM}},
	}
}

type SetupRole struct {
	Key          string
	Capabilities []Capability
	Required     bool
	SortOrder    int
}

type GoalPolicy struct {
	Goal  planning.Goal
	Roles []SetupRole
}

type Weights struct {
	GoalMatch       int
	BudgetMatch     int
	SpaceMatch      int
	ExperienceMatch int
	PreferenceMatch int
	Quality         int
	Value           int
	Durability      int
	Compatibility   int
	Portability     int
	Noise           int
}

type Config struct {
	PolicyVersion        string
	Weights              Weights
	PriorityBoostPercent int
	MaximumSetupItems    int
	CandidatesPerSlot    int
	OptionalSlotBonus    int
	Goals                []GoalPolicy
}

type ScoreBreakdown struct {
	GoalMatch       int `json:"goal_match"`
	BudgetMatch     int `json:"budget_match"`
	SpaceMatch      int `json:"space_match"`
	ExperienceMatch int `json:"experience_match"`
	PreferenceMatch int `json:"preference_match"`
	Quality         int `json:"quality"`
	Value           int `json:"value"`
	Durability      int `json:"durability"`
	Compatibility   int `json:"compatibility"`
	Portability     int `json:"portability"`
	Noise           int `json:"noise"`
}

type Reason struct {
	Code      string
	Message   string
	Dimension string
	Score     int
}

type RankedProduct struct {
	Candidate      CandidateSnapshot
	ObjectiveScore int
	Breakdown      ScoreBreakdown
	Reasons        []Reason
}

type RecommendedItem struct {
	Rank           int
	Product        RankedProduct
	Quantity       int
	UnitPriceMinor int64
}

type AlternativeType string

const (
	AlternativeCheaper AlternativeType = "cheaper"
	AlternativePremium AlternativeType = "premium"
)

type Alternative struct {
	ForProductID         catalog.ProductID
	Type                 AlternativeType
	Product              RankedProduct
	PriceDifferenceMinor int64
}

type RejectedProduct struct {
	Candidate CandidateSnapshot
	Code      string
	Message   string
}

type ResultStatus string

const (
	ResultComplete           ResultStatus = "complete"
	ResultNoSuitableProducts ResultStatus = "no_suitable_products"
)

type Result struct {
	Status           ResultStatus
	PolicyVersion    string
	EngineVersion    string
	InputFingerprint string
	ObjectiveScore   int
	Breakdown        ScoreBreakdown
	Selected         []RecommendedItem
	TotalCost        catalog.Money
	Alternatives     []Alternative
	Rejected         []RejectedProduct
	Ranked           []RankedProduct
}
