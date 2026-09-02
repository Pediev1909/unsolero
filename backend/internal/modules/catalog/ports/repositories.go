package ports

import (
	"context"
	"errors"

	"rigmark/internal/modules/catalog/domain"
)

var ErrNotFound = errors.New("catalog entity not found")

type ProductFilter struct {
	ProductIDs    []domain.ProductID
	CategorySlug  string
	BrandSlug     string
	Search        string
	Sort          string
	ExcludeSlug   string
	MinPriceMinor *int64
	MaxPriceMinor *int64
	// HasOffer keeps only products with a servable affiliate offer, under the
	// same conditions the vendor button and the redirect apply.
	HasOffer bool
	Offset   int
	Limit    int
}

type ProductPage struct {
	Products []domain.Product
	Total    int
}

type CategoryRepository interface {
	ListActiveCategories(context.Context) ([]domain.Category, error)
	GetActiveCategoryBySlug(context.Context, string) (domain.Category, error)
}

type BrandRepository interface {
	ListActiveBrands(context.Context) ([]domain.Brand, error)
	// ListActiveBrandsInCategory returns only brands that have a published
	// product in the given category. The filter on a category page listed
	// every brand in the catalog, so picking one that sells nothing in that
	// category produced an empty result the interface had invited.
	ListActiveBrandsInCategory(context.Context, string) ([]domain.Brand, error)
	GetActiveBrandBySlug(context.Context, string) (domain.Brand, error)
}

type ProductRepository interface {
	GetPublishedBySlug(context.Context, string) (domain.Product, error)
	ListPublished(context.Context, ProductFilter) ([]domain.Product, error)
	SearchPublished(context.Context, ProductFilter) (ProductPage, error)
}

// PriceRecordRepository reads the dated price history of one product out of its
// immutable fact revisions.
//
// It is deliberately separate from ProductRepository and takes one product at a
// time: this is a detail-page read, and no listing wants ten historic prices
// per card. A caller asking for a page of products must not be able to reach
// it by accident.
type PriceRecordRepository interface {
	ListPriceRecord(context.Context, domain.ProductID) ([]domain.PriceObservation, error)
}
