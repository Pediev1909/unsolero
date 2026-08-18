package ports

import (
	"context"
	"errors"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

var ErrProductNotFound = errors.New("wishlist product not found")

type WishlistRepository interface {
	ListProductIDs(context.Context, identity.UserID, int, int) (WishlistPage, error)
	Save(context.Context, identity.UserID, catalog.ProductID) error
	Delete(context.Context, identity.UserID, catalog.ProductID) error
}

type WishlistPage struct {
	ProductIDs []catalog.ProductID
	Total      int
}
