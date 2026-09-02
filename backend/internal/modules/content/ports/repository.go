package ports

import (
	"context"
	"errors"

	"rigmark/internal/modules/content/domain"
)

var ErrNotFound = errors.New("editorial content not found")

type Filter struct {
	Types        []domain.ContentType
	CategorySlug string
	// ProductSlug narrows the list to entries whose product references include
	// this product. It is how a product page finds the comparisons and guides
	// it appears in.
	ProductSlug string
	AuthorSlug  string
	ExcludeID   string
	Limit       int
}

type Repository interface {
	ListPublished(context.Context, Filter) ([]domain.Summary, error)
	GetAuthorBySlug(context.Context, string) (domain.Author, error)
	GetPublishedBySlug(context.Context, string) (domain.Entry, error)
	ListSitemapEntries(context.Context) ([]domain.SitemapEntry, error)
}
