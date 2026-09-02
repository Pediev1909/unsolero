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

func (stub *repositoryStub) GetAuthorBySlug(_ context.Context, slug string) (domain.Author, error) {
	return domain.Author{Name: "Andon Pediev", Slug: slug, Bio: "Builds and runs UNSOLERO."}, nil
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
	if _, err := service.List(context.Background(), ListQuery{Section: "guides", CategorySlug: "adjustable-dumbbells", Limit: 12}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repository.filter.Types) != 2 || repository.filter.Types[0] != domain.ContentTypeGuide ||
		repository.filter.Types[1] != domain.ContentTypeBuyingGuide {
		t.Fatalf("guide filter types = %#v", repository.filter.Types)
	}
}

// The stacks hub lists one type. The section is the URL's plural and the type
// is the row's singular; the mapping is the only place the two meet.
func TestListMapsStacksSectionToStackType(t *testing.T) {
	repository := &repositoryStub{}
	service, _ := NewService(repository, nil, "https://rigmark.example")
	if _, err := service.List(context.Background(), ListQuery{Section: "stacks", Limit: 24}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(repository.filter.Types) != 1 || repository.filter.Types[0] != domain.ContentTypeStack {
		t.Fatalf("stack filter types = %#v", repository.filter.Types)
	}
}

func TestListRejectsUnknownSections(t *testing.T) {
	service, _ := NewService(&repositoryStub{}, nil, "https://rigmark.example")
	if _, err := service.List(context.Background(), ListQuery{Section: "generated", Limit: 12}); err == nil {
		t.Fatal("List() expected invalid section error")
	}
}

func TestListPassesProductSlugToRepository(t *testing.T) {
	repository := &repositoryStub{}
	service, _ := NewService(repository, nil, "https://rigmark.example")
	if _, err := service.List(context.Background(), ListQuery{ProductSlug: "mailchimp-standard", Limit: 12}); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.filter.ProductSlug != "mailchimp-standard" || repository.filter.Types != nil {
		t.Fatalf("product filter = %#v", repository.filter)
	}
}

// The product slug is read from the URL, so it gets the same treatment as the
// category slug: anything that is not a slug is refused before it reaches SQL.
func TestListRejectsMalformedProductSlug(t *testing.T) {
	service, _ := NewService(&repositoryStub{}, nil, "https://rigmark.example")
	for _, slug := range []string{"Mailchimp Standard", "mailchimp_standard", "-mailchimp", "a--b", "x' OR 1=1"} {
		if _, err := service.List(context.Background(), ListQuery{ProductSlug: slug, Limit: 12}); err == nil {
			t.Fatalf("List(product=%q) expected invalid query error", slug)
		}
	}
}
