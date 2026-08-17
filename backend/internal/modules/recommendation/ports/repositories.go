package ports

import (
	"context"
	"errors"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/domain"
)

var ErrNotFound = errors.New("recommendation resource not found")
var ErrConflict = errors.New("recommendation resource conflict")
var ErrSeparationOfDuties = errors.New("recommendation policy separation of duties violation")

type Draft struct {
	CurrentStep         int
	Goal                *planning.Goal
	Experience          *planning.ExperienceLevel
	BudgetMinor         *int64
	Currency            *string
	AvailableSpace      *domain.AvailableSpace
	ExistingEquipment   []domain.ExistingEquipment
	TrainingPreferences []domain.TrainingPreference
	Priorities          []domain.Priority
	FreeText            string
	UpdatedAt           time.Time
}

type SavedResult struct {
	RecommendationID domain.RecommendationID
	SetupID          planning.SetupID
	SetupName        string
}

type PersistedResult struct {
	RecommendationID domain.RecommendationID
	SetupID          planning.SetupID
	SetupName        string
	Input            domain.Input
	Result           domain.Result
	CreatedAt        time.Time
}

type SetupSummary struct {
	ID             planning.SetupID
	Name           string
	ItemCount      int
	TotalCostMinor int64
	Currency       string
	ObjectiveScore int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Repository interface {
	GetDraft(context.Context, identity.UserID) (Draft, error)
	SaveDraft(context.Context, identity.UserID, Draft) (Draft, error)
	DeleteDraft(context.Context, identity.UserID) error
	SaveResult(context.Context, identity.UserID, domain.Input, domain.Result, []domain.CandidateSnapshot) (SavedResult, error)
	ListSetups(context.Context, identity.UserID) ([]SetupSummary, error)
	GetResultBySetupID(context.Context, identity.UserID, planning.SetupID) (PersistedResult, error)
	RenameSetup(context.Context, identity.UserID, planning.SetupID, string) error
	DeleteSetup(context.Context, identity.UserID, planning.SetupID) error
}

type PolicyRepository interface {
	ActivePolicy(context.Context) (domain.Policy, error)
	ListPolicies(context.Context) ([]domain.PolicySummary, error)
	TransitionPolicy(context.Context, identity.UserID, string, domain.PolicyWorkflowStatus, string) error
}

type CatalogRepository interface {
	ListPublished(context.Context, catalogports.ProductFilter) ([]catalog.Product, error)
	ListByIDs(context.Context, []catalog.ProductID) ([]catalog.Product, error)
}
