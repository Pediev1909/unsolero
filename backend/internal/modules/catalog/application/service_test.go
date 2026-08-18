package application

import (
	"testing"

	"rigmark/internal/modules/catalog/domain"
)

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
