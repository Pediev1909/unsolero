package domain

import (
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
)

type SessionID string
type RecommendationID string
type ItemID string

type Session struct {
	ID              SessionID
	UserID          *identity.UserID
	ProfileID       *planning.ProfileID
	AnonymousID     *string
	Status          string
	PrimaryGoal     planning.Goal
	ExperienceLevel planning.ExperienceLevel
	BudgetMinor     int64
	Currency        string
	SpaceLengthMM   *int64
	SpaceWidthMM    *int64
	SpaceHeightMM   *int64
	ApartmentLiving bool
	StartedAt       time.Time
	CompletedAt     *time.Time
	ExpiresAt       *time.Time
}

type Recommendation struct {
	ID                RecommendationID
	SessionID         SessionID
	PolicyVersion     string
	EngineVersion     string
	Status            string
	ObjectiveScore    int16
	TotalPriceMinor   int64
	Currency          string
	ResultFingerprint string
	Items             []Item
	CreatedAt         time.Time
}

type Item struct {
	ID             ItemID
	ProductID      catalog.ProductID
	Type           string
	Rank           int16
	Quantity       int16
	UnitPriceMinor int64
	Currency       string
	ObjectiveScore int16
	ReasonCode     string
	ReasonSummary  string
	RejectionCode  *string
	CreatedAt      time.Time
}
