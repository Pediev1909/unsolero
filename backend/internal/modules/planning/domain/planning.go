package domain

import (
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type ProfileID string
type SetupID string
type SetupItemID string
type WishlistID string

type ExperienceLevel string

const (
	ExperienceBeginner     ExperienceLevel = "beginner"
	ExperienceIntermediate ExperienceLevel = "intermediate"
	ExperienceAdvanced     ExperienceLevel = "advanced"
)

type Goal string

const (
	GoalBuildMuscle    Goal = "build_muscle"
	GoalStrength       Goal = "strength"
	GoalGeneralFitness Goal = "general_fitness"
	GoalWeightLoss     Goal = "weight_loss"
	GoalMobility       Goal = "mobility"
)

type Profile struct {
	ID                  ProfileID
	UserID              identity.UserID
	DisplayName         string
	ExperienceLevel     ExperienceLevel
	PrimaryGoal         Goal
	BudgetMinor         *int64
	Currency            *string
	SpaceLengthMM       *int64
	SpaceWidthMM        *int64
	SpaceHeightMM       *int64
	ApartmentLiving     bool
	NoiseToleranceScore *int16
	CountryCode         *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Setup struct {
	ID                     SetupID
	UserID                 identity.UserID
	SourceRecommendationID *string
	Name                   string
	Description            string
	Currency               string
	Items                  []SetupItem
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SetupItem struct {
	ID              SetupItemID
	ProductID       catalog.ProductID
	MerchantOfferID *string
	Quantity        int16
	PurchaseStatus  string
	PaidPriceMinor  *int64
	Currency        *string
	AddedAt         time.Time
	UpdatedAt       time.Time
}

type WishlistEntry struct {
	ID        WishlistID
	UserID    identity.UserID
	ProductID catalog.ProductID
	Priority  int16
	Note      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
