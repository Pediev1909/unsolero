package ports

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var ErrProductNotFound = errors.New("wishlist product not found")

type WishlistRepository interface {
	ListProductIDs(context.Context, identity.UserID) ([]catalog.ProductID, error)
	Save(context.Context, identity.UserID, catalog.ProductID) error
	Delete(context.Context, identity.UserID, catalog.ProductID) error
}
