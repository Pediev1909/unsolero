package ports

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/evidence/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var (
	ErrNotFound             = errors.New("evidence resource not found")
	ErrConflict             = errors.New("evidence state conflict")
	ErrSeparationOfDuties   = errors.New("a revision author cannot approve their own work")
	ErrIncompleteProvenance = errors.New("revision provenance is incomplete, unverified, or stale")
)

type Repository interface {
	CreateSource(context.Context, identity.UserID, domain.SourceInput) (domain.Source, error)
	ReviewSource(context.Context, identity.UserID, string, domain.ReviewStatus, string) (domain.Source, error)
	CreateObservation(context.Context, identity.UserID, domain.ObservationInput) (domain.Observation, error)
	CreateRevision(context.Context, identity.UserID, domain.RevisionInput) (domain.Revision, error)
	TransitionRevision(context.Context, identity.UserID, string, domain.WorkflowStatus, string) (domain.Revision, error)
	PublishRevision(context.Context, identity.UserID, string) (domain.Revision, error)
	GetProductGovernance(context.Context, catalog.ProductID) (domain.ProductGovernance, error)
	ListProductGovernance(context.Context, int, int) ([]domain.ProductGovernance, int64, error)
}
