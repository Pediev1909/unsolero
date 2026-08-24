package application

import (
	"context"
	"errors"
	"strings"

	"rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
)

var ErrInvalidQuery = errors.New("invalid catalog query")

const (
	defaultPageSize = 12
	maximumPageSize = 48
	maximumPage     = 10_000
)

// MaximumPageSize lets a caller build a query it knows will be accepted. Search
// rejects an over-sized page size as an invalid query rather than clamping it,
// and a caller that treats an error as "no results" then renders an empty page
// with no indication anything went wrong.
const MaximumPageSize = maximumPageSize

type Query struct {
	ProductIDs    []domain.ProductID
	Search        string
	CategorySlug  string
	BrandSlug     string
	Sort          string
	MinPriceMinor *int64
	MaxPriceMinor *int64
	Page          int
	PageSize      int
}

type Page struct {
	Products   []domain.Product
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

type ProductDetail struct {
	Product        domain.Product
	Alternatives   []domain.Product
	Strengths      []domain.SuitabilityInsight
	Considerations []domain.SuitabilityInsight
	UseCases       []domain.SuitabilityInsight
}

type Repository interface {
	ports.ProductRepository
	ports.CategoryRepository
	ports.BrandRepository
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (service *Service) Search(ctx context.Context, query Query) (Page, error) {
	normalized, err := normalizeQuery(query)
	if err != nil {
		return Page{}, err
	}
	result, err := service.repository.SearchPublished(ctx, ports.ProductFilter{
		ProductIDs:    normalized.ProductIDs,
		CategorySlug:  normalized.CategorySlug,
		BrandSlug:     normalized.BrandSlug,
		Search:        normalized.Search,
		Sort:          normalized.Sort,
		MinPriceMinor: normalized.MinPriceMinor,
		MaxPriceMinor: normalized.MaxPriceMinor,
		Offset:        (normalized.Page - 1) * normalized.PageSize,
		Limit:         normalized.PageSize,
	})
	if err != nil {
		return Page{}, err
	}
	totalPages := 0
	if result.Total > 0 {
		totalPages = (result.Total + normalized.PageSize - 1) / normalized.PageSize
	}
	return Page{
		Products: result.Products, Page: normalized.Page,
		PageSize: normalized.PageSize, Total: result.Total, TotalPages: totalPages,
	}, nil
}

func (service *Service) GetProduct(ctx context.Context, slug string) (ProductDetail, error) {
	if !validSlug(slug) {
		return ProductDetail{}, ports.ErrNotFound
	}
	product, err := service.repository.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return ProductDetail{}, err
	}
	alternatives, err := service.repository.SearchPublished(ctx, ports.ProductFilter{
		CategorySlug: product.CategorySlug, ExcludeSlug: slug, Sort: "featured", Limit: 4,
	})
	if err != nil {
		return ProductDetail{}, err
	}
	return ProductDetail{
		Product: product, Alternatives: alternatives.Products,
		Strengths: product.Strengths(), Considerations: product.Considerations(),
		UseCases: product.UseCases(),
	}, nil
}

func (service *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return service.repository.ListActiveCategories(ctx)
}

func (service *Service) GetCategory(ctx context.Context, slug string) (domain.Category, error) {
	if !validSlug(slug) {
		return domain.Category{}, ports.ErrNotFound
	}
	return service.repository.GetActiveCategoryBySlug(ctx, slug)
}

func (service *Service) ListBrands(ctx context.Context) ([]domain.Brand, error) {
	return service.repository.ListActiveBrands(ctx)
}

// ListBrandsInCategory narrows the brand filter to what the category actually
// holds. An unknown slug returns nothing rather than everything: a filter that
// silently ignores the constraint it was given is worse than one that returns
// an empty list, because the reader cannot tell it was ignored.
func (service *Service) ListBrandsInCategory(ctx context.Context, categorySlug string) ([]domain.Brand, error) {
	if !validSlug(categorySlug) {
		return []domain.Brand{}, nil
	}
	return service.repository.ListActiveBrandsInCategory(ctx, categorySlug)
}

func (service *Service) GetBrand(ctx context.Context, slug string) (domain.Brand, error) {
	if !validSlug(slug) {
		return domain.Brand{}, ports.ErrNotFound
	}
	return service.repository.GetActiveBrandBySlug(ctx, slug)
}

func normalizeQuery(query Query) (Query, error) {
	query.Search = strings.TrimSpace(query.Search)
	query.CategorySlug = strings.TrimSpace(query.CategorySlug)
	query.BrandSlug = strings.TrimSpace(query.BrandSlug)
	query.Sort = strings.TrimSpace(query.Sort)
	if len(query.ProductIDs) > maximumPageSize {
		return Query{}, ErrInvalidQuery
	}
	seenProductIDs := make(map[domain.ProductID]bool, len(query.ProductIDs))
	for _, productID := range query.ProductIDs {
		if productID == "" || seenProductIDs[productID] {
			return Query{}, ErrInvalidQuery
		}
		seenProductIDs[productID] = true
	}
	if len(query.Search) > 100 || query.CategorySlug != "" && !validSlug(query.CategorySlug) ||
		query.BrandSlug != "" && !validSlug(query.BrandSlug) {
		return Query{}, ErrInvalidQuery
	}
	if query.Sort == "" {
		query.Sort = "featured"
	}
	validSort := map[string]bool{
		"featured": true, "name_asc": true, "price_asc": true,
		"price_desc": true, "quality_desc": true, "value_desc": true,
	}
	if !validSort[query.Sort] {
		return Query{}, ErrInvalidQuery
	}
	if query.MinPriceMinor != nil && *query.MinPriceMinor < 0 ||
		query.MaxPriceMinor != nil && *query.MaxPriceMinor < 0 ||
		query.MinPriceMinor != nil && query.MaxPriceMinor != nil && *query.MinPriceMinor > *query.MaxPriceMinor {
		return Query{}, ErrInvalidQuery
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Page > maximumPage {
		return Query{}, ErrInvalidQuery
	}
	if query.PageSize <= 0 {
		query.PageSize = defaultPageSize
	}
	if query.PageSize > maximumPageSize {
		return Query{}, ErrInvalidQuery
	}
	return query, nil
}

func validSlug(value string) bool {
	if value == "" || len(value) > 160 || strings.HasPrefix(value, "-") ||
		strings.HasSuffix(value, "-") || strings.Contains(value, "--") {
		return false
	}
	return strings.IndexFunc(value, func(character rune) bool {
		return (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '-'
	}) == -1
}
