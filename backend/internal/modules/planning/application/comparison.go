package application

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

var ErrInvalidComparison = errors.New("comparison must contain up to four unique products")

type ComparisonService struct {
	repository ports.ComparisonRepository
}

func NewComparisonService(repository ports.ComparisonRepository) *ComparisonService {
	return &ComparisonService{repository: repository}
}

func (service *ComparisonService) List(
	ctx context.Context,
	userID identity.UserID,
) ([]catalog.ProductID, error) {
	return service.repository.ListProductIDs(ctx, userID)
}

func (service *ComparisonService) Replace(
	ctx context.Context,
	userID identity.UserID,
	productIDs []catalog.ProductID,
) error {
	if len(productIDs) > 4 {
		return ErrInvalidComparison
	}
	seen := make(map[catalog.ProductID]bool, len(productIDs))
	for _, productID := range productIDs {
		if productID == "" || seen[productID] {
			return ErrInvalidComparison
		}
		seen[productID] = true
	}
	return service.repository.Replace(ctx, userID, productIDs)
}
