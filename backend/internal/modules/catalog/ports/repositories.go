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
	Offset        int
	Limit         int
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
	GetActiveBrandBySlug(context.Context, string) (domain.Brand, error)
}

type ProductRepository interface {
	GetPublishedBySlug(context.Context, string) (domain.Product, error)
	ListPublished(context.Context, ProductFilter) ([]domain.Product, error)
	SearchPublished(context.Context, ProductFilter) (ProductPage, error)
}
