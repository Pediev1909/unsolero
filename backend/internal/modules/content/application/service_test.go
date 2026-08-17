package application

import (
	"context"
	"testing"

	"rigmark/internal/modules/content/domain"
	"rigmark/internal/modules/content/ports"
)

type repositoryStub struct {
	filter ports.Filter
}

func (stub *repositoryStub) ListPublished(_ context.Context, filter ports.Filter) ([]domain.Summary, error) {
	stub.filter = filter
	return []domain.Summary{}, nil
}
func (*repositoryStub) GetPublishedBySlug(context.Context, string) (domain.Entry, error) {
	return domain.Entry{}, ports.ErrNotFound
}
func (*repositoryStub) ListSitemapEntries(context.Context) ([]domain.SitemapEntry, error) {
	return nil, nil
}

func TestListMapsGuideSectionToEditorialTypes(t *testing.T) {
	repository := &repositoryStub{}
	service, err := NewService(repository, nil, "https://rigmark.example")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	if _, err := service.List(context.Background(), "guides", "adjustable-dumbbells", 12); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repository.filter.Types) != 2 || repository.filter.Types[0] != domain.ContentTypeGuide ||
		repository.filter.Types[1] != domain.ContentTypeBuyingGuide {
		t.Fatalf("guide filter types = %#v", repository.filter.Types)
	}
}

func TestListRejectsUnknownSections(t *testing.T) {
	service, _ := NewService(&repositoryStub{}, nil, "https://rigmark.example")
	if _, err := service.List(context.Background(), "generated", "", 12); err == nil {
		t.Fatal("List() expected invalid section error")
	}
}
