package application

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/planning/ports"
)

var ErrInvalidPagination = errors.New("wishlist pagination is invalid")

type WishlistService struct {
	repository ports.WishlistRepository
}

func NewWishlistService(repository ports.WishlistRepository) *WishlistService {
	return &WishlistService{repository: repository}
}

func (service *WishlistService) List(
	ctx context.Context,
	userID identity.UserID,
	page int,
	pageSize int,
) (ports.WishlistPage, error) {
	if page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		return ports.WishlistPage{}, ErrInvalidPagination
	}
	return service.repository.ListProductIDs(ctx, userID, pageSize, (page-1)*pageSize)
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
