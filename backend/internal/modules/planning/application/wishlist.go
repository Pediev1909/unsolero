package application

import (
	"context"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

type WishlistService struct {
	repository ports.WishlistRepository
}

func NewWishlistService(repository ports.WishlistRepository) *WishlistService {
	return &WishlistService{repository: repository}
}

func (service *WishlistService) List(
	ctx context.Context,
	userID identity.UserID,
) ([]catalog.ProductID, error) {
	return service.repository.ListProductIDs(ctx, userID)
}

func (service *WishlistService) Save(
	ctx context.Context,
	userID identity.UserID,
	productID catalog.ProductID,
) error {
	return service.repository.Save(ctx, userID, productID)
}

func (service *WishlistService) Delete(
	ctx context.Context,
	userID identity.UserID,
	productID catalog.ProductID,
) error {
	return service.repository.Delete(ctx, userID, productID)
}
