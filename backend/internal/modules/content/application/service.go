package application

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"

	catalogports "rigmark/internal/modules/catalog/ports"
	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

var ErrInvalidQuery = errors.New("invalid editorial query")

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Catalog interface {
	SearchPublished(context.Context, catalogports.ProductFilter) (catalogports.ProductPage, error)
}

type Service struct {
	repository ports.Repository
	catalog    Catalog
	siteURL    *url.URL
}

func NewService(repository ports.Repository, catalog Catalog, siteURL string) (*Service, error) {
	parsed, err := url.Parse(strings.TrimRight(siteURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("public site URL must be absolute")
	}
	return &Service{repository: repository, catalog: catalog, siteURL: parsed}, nil
}

func (service *Service) List(ctx context.Context, section, categorySlug string, limit int) ([]domain.Summary, error) {
	types, err := sectionTypes(section)
	if err != nil || categorySlug != "" && !slugPattern.MatchString(categorySlug) || limit < 1 || limit > 24 {
		return nil, ErrInvalidQuery
	}
	return service.repository.ListPublished(ctx, ports.Filter{
		Types: types, CategorySlug: categorySlug, Limit: limit,
	})
}

func (service *Service) Get(ctx context.Context, slug string) (domain.Entry, error) {
	if !slugPattern.MatchString(slug) {
		return domain.Entry{}, ports.ErrNotFound
	}
	entry, err := service.repository.GetPublishedBySlug(ctx, slug)
	if err != nil {
		return domain.Entry{}, err
	}
	entry.Path = entry.Type.Path(entry.Slug)
	if entry.CanonicalURL == "" {
		entry.CanonicalURL = service.siteURL.ResolveReference(&url.URL{Path: entry.Path}).String()
	}
	if len(entry.ProductIDs) > 0 && service.catalog != nil {
		products, err := service.catalog.SearchPublished(ctx, catalogports.ProductFilter{
			ProductIDs: entry.ProductIDs, Limit: len(entry.ProductIDs),
		})
		if err != nil {
			return domain.Entry{}, err
		}
		entry.RelatedProducts = products.Products
	}
	if err := entry.Validate(); err != nil {
		return domain.Entry{}, err
	}
	return entry, nil
}

// Author returns one author with everything they have published. The byline on
// an entry is only worth as much as the page behind it: a reader deciding
// whether to trust a ranking, and a search engine weighing who produced it,
// both need somewhere to look the person up.
func (service *Service) Author(ctx context.Context, slug string) (domain.Author, []domain.Summary, error) {
	author, err := service.repository.GetAuthorBySlug(ctx, slug)
	if err != nil {
		return domain.Author{}, nil, err
	}
	entries, err := service.repository.ListPublished(ctx, ports.Filter{
		AuthorSlug: author.Slug, Limit: 100,
	})
	if err != nil {
		return domain.Author{}, nil, err
	}
	return author, entries, nil
}

func (service *Service) Sitemap(ctx context.Context) ([]domain.SitemapEntry, error) {
	return service.repository.ListSitemapEntries(ctx)
}

func (service *Service) AbsoluteURL(path string) string {
	return service.siteURL.ResolveReference(&url.URL{Path: path}).String()
}

func sectionTypes(section string) ([]domain.ContentType, error) {
	switch strings.TrimSpace(section) {
	case "", "all":
		return nil, nil
	case "articles":
		return []domain.ContentType{domain.ContentTypeArticle}, nil
	case "guides":
		return []domain.ContentType{domain.ContentTypeGuide, domain.ContentTypeBuyingGuide}, nil
	case "comparisons":
		return []domain.ContentType{domain.ContentTypeComparison}, nil
	default:
		return nil, ErrInvalidQuery
	}
}
