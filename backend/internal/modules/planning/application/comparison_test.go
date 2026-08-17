package application

import (
	"context"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type comparisonRepositoryStub struct{ replaced []catalog.ProductID }

func (repository *comparisonRepositoryStub) ListProductIDs(context.Context, identity.UserID) ([]catalog.ProductID, error) {
	return append([]catalog.ProductID(nil), repository.replaced...), nil
}
func (repository *comparisonRepositoryStub) Replace(_ context.Context, _ identity.UserID, ids []catalog.ProductID) error {
	repository.replaced = append([]catalog.ProductID(nil), ids...)
	return nil
}

func TestComparisonServiceReplacesOrderedSelection(t *testing.T) {
	repository := &comparisonRepositoryStub{}
	service := NewComparisonService(repository)
	ids := []catalog.ProductID{"one", "two", "three", "four"}
	if err := service.Replace(context.Background(), "user", ids); err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	for index := range ids {
		if repository.replaced[index] != ids[index] {
			t.Fatalf("selection order = %#v, want %#v", repository.replaced, ids)
		}
	}
}

func TestComparisonServiceRejectsInvalidSelections(t *testing.T) {
	service := NewComparisonService(&comparisonRepositoryStub{})
	cases := [][]catalog.ProductID{
		{"one", "two", "three", "four", "five"},
		{"one", "one"},
		{""},
	}
	for _, ids := range cases {
		if err := service.Replace(context.Background(), "user", ids); err != ErrInvalidComparison {
			t.Fatalf("Replace(%#v) error = %v, want ErrInvalidComparison", ids, err)
		}
	}
}
