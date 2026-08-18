package merchant

import (
	"context"
	"errors"
	"testing"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

func TestMissingProviderFailsClosed(t *testing.T) {
	provider := NewRegistry().Select("awin")
	if provider.Key() != "awin" {
		t.Fatalf("provider key = %q", provider.Key())
	}
	if _, err := provider.FetchOffers(context.Background(), domain.ProviderConfiguration{}, nil); !errors.Is(err, ports.ErrProviderDisabled) {
		t.Fatalf("FetchOffers error = %v, want disabled", err)
	}
}
