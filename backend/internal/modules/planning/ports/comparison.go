package ports

import (
	"context"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type ComparisonRepository interface {
	ListProductIDs(context.Context, identity.UserID) ([]catalog.ProductID, error)
	Replace(context.Context, identity.UserID, []catalog.ProductID) error
}
