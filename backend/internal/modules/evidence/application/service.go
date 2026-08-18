package application

import (
	"context"
	"errors"
	"strings"
	"time"

	catalog "rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/evidence/domain"
	"rigmark/internal/modules/evidence/ports"
	identity "rigmark/internal/modules/identity/domain"
)

var ErrInvalidInput = errors.New("evidence input is invalid")

type Service struct {
	repository ports.Repository
	now        func() time.Time
}

func NewService(repository ports.Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (service *Service) CreateSource(ctx context.Context, actor identity.UserID, input domain.SourceInput) (domain.Source, error) {
	input.Title, input.Publisher = strings.TrimSpace(input.Title), strings.TrimSpace(input.Publisher)
	if input.URL != nil {
		value := strings.TrimSpace(*input.URL)
		input.URL = &value
	}
	if actor == "" || input.Validate() != nil {
		return domain.Source{}, ErrInvalidInput
	}
	return service.repository.CreateSource(ctx, actor, input)
}

func (service *Service) ReviewSource(ctx context.Context, actor identity.UserID, id string, status domain.ReviewStatus, note string) (domain.Source, error) {
	note = strings.TrimSpace(note)
	if actor == "" || id == "" || (status != domain.ReviewVerified && status != domain.ReviewRejected) || len(note) > 2000 {
		return domain.Source{}, ErrInvalidInput
	}
	return service.repository.ReviewSource(ctx, actor, id, status, note)
}

func (service *Service) CreateObservation(ctx context.Context, actor identity.UserID, input domain.ObservationInput) (domain.Observation, error) {
	input.Notes = strings.TrimSpace(input.Notes)
	if actor == "" || input.Validate(service.now()) != nil {
		return domain.Observation{}, ErrInvalidInput
	}
	return service.repository.CreateObservation(ctx, actor, input)
}

func (service *Service) CreateRevision(ctx context.Context, actor identity.UserID, input domain.RevisionInput) (domain.Revision, error) {
	if actor == "" || input.Validate() != nil {
		return domain.Revision{}, ErrInvalidInput
	}
	return service.repository.CreateRevision(ctx, actor, input)
}

func (service *Service) Submit(ctx context.Context, actor identity.UserID, revisionID string) (domain.Revision, error) {
	return service.transition(ctx, actor, revisionID, domain.WorkflowInReview, "")
}

func (service *Service) Approve(ctx context.Context, actor identity.UserID, revisionID, note string) (domain.Revision, error) {
	return service.transition(ctx, actor, revisionID, domain.WorkflowApproved, note)
}

func (service *Service) Reject(ctx context.Context, actor identity.UserID, revisionID, note string) (domain.Revision, error) {
	if strings.TrimSpace(note) == "" {
		return domain.Revision{}, ErrInvalidInput
	}
	return service.transition(ctx, actor, revisionID, domain.WorkflowRejected, note)
}

func (service *Service) transition(ctx context.Context, actor identity.UserID, revisionID string, status domain.WorkflowStatus, note string) (domain.Revision, error) {
	note = strings.TrimSpace(note)
	if actor == "" || revisionID == "" || len(note) > 2000 {
		return domain.Revision{}, ErrInvalidInput
	}
	return service.repository.TransitionRevision(ctx, actor, revisionID, status, note)
}

func (service *Service) Publish(ctx context.Context, actor identity.UserID, revisionID string) (domain.Revision, error) {
	if actor == "" || revisionID == "" {
		return domain.Revision{}, ErrInvalidInput
	}
	return service.repository.PublishRevision(ctx, actor, revisionID)
}

func (service *Service) GetProduct(ctx context.Context, productID catalog.ProductID) (domain.ProductGovernance, error) {
	if productID == "" {
		return domain.ProductGovernance{}, ErrInvalidInput
	}
	return service.repository.GetProductGovernance(ctx, productID)
}

func (service *Service) ListProducts(ctx context.Context, page, pageSize int) ([]domain.ProductGovernance, int64, error) {
	if page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		return nil, 0, ErrInvalidInput
	}
	return service.repository.ListProductGovernance(ctx, pageSize, (page-1)*pageSize)
}
