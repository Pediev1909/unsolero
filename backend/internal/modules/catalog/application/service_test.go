package application

import (
	"context"
	"testing"

	"rigmark/internal/modules/catalog/domain"
	"rigmark/internal/modules/catalog/ports"
)

// filterCapturingRepository records the filter Search hands the repository.
// Embedding the interface leaves every other method unimplemented, which is
// fine: Search calls exactly one.
type filterCapturingRepository struct {
	Repository
	filter ports.ProductFilter
}

func (repository *filterCapturingRepository) SearchPublished(_ context.Context, filter ports.ProductFilter) (ports.ProductPage, error) {
	repository.filter = filter
	return ports.ProductPage{}, nil
}

// The live-offer filter has to reach the query that computes the total, not a
// loop over the returned page; otherwise the response says "12 products" and
// draws four.
func TestSearchThreadsHasOfferIntoTheListingFilter(t *testing.T) {
	repository := &filterCapturingRepository{}
	if _, err := NewService(repository).Search(context.Background(), Query{HasOffer: true, Page: 2, PageSize: 12}); err != nil {
		t.Fatalf("Search() returned an error: %v", err)
	}
	if !repository.filter.HasOffer || repository.filter.Offset != 12 || repository.filter.Limit != 12 {
		t.Fatalf("repository filter = %#v, want HasOffer with the page's offset and limit", repository.filter)
	}
	if _, err := NewService(repository).Search(context.Background(), Query{}); err != nil {
		t.Fatalf("Search() returned an error: %v", err)
	}
	if repository.filter.HasOffer {
		t.Fatal("HasOffer must be off unless the query asked for it")
	}
}

func TestNormalizeQueryAppliesDefaults(t *testing.T) {
	query, err := normalizeQuery(Query{})
	if err != nil {
		t.Fatalf("normalize query: %v", err)
	}
	if query.Page != 1 || query.PageSize != 12 || query.Sort != "featured" {
		t.Fatalf("unexpected defaults: %#v", query)
	}
}

func TestNormalizeQueryRejectsInvertedPriceRange(t *testing.T) {
	minimum := int64(20000)
	maximum := int64(10000)
	_, err := normalizeQuery(Query{MinPriceMinor: &minimum, MaxPriceMinor: &maximum})
	if err != ErrInvalidQuery {
		t.Fatalf("expected ErrInvalidQuery, got %v", err)
	}
}

func TestNormalizeQueryRejectsDuplicateProductIDs(t *testing.T) {
	_, err := normalizeQuery(Query{ProductIDs: []domain.ProductID{"one", "one"}})
	if err != ErrInvalidQuery {
		t.Fatalf("expected ErrInvalidQuery, got %v", err)
	}
}

func TestNormalizeQueryRejectsExcessivePage(t *testing.T) {
	if _, err := normalizeQuery(Query{Page: maximumPage + 1, PageSize: 48}); err != ErrInvalidQuery {
		t.Fatalf("normalizeQuery() error = %v, want %v", err, ErrInvalidQuery)
	}
}

func TestValidSlug(t *testing.T) {
	for _, value := range []string{"adjustable-dumbbells", "demo-product-20"} {
		if !validSlug(value) {
			t.Fatalf("expected valid slug %q", value)
		}
	}
	for _, value := range []string{"", "Invalid", "../product", "two--hyphens"} {
		if validSlug(value) {
			t.Fatalf("expected invalid slug %q", value)
		}
	}
}
