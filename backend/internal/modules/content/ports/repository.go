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
	ExcludeID    string
	Limit        int
}

type Repository interface {
	ListPublished(context.Context, Filter) ([]domain.Summary, error)
	GetPublishedBySlug(context.Context, string) (domain.Entry, error)
	ListSitemapEntries(context.Context) ([]domain.SitemapEntry, error)
}
