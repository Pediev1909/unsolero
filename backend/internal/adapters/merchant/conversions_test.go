package merchant

import (
	"context"
	"errors"
	"testing"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

func TestUnknownConversionProviderFailsClosed(t *testing.T) {
	adapter := NewConversionRegistry().Select("awin")
	if err := adapter.ValidateConversionConfiguration(context.Background(), domain.ProviderConfiguration{}); !errors.Is(err, ports.ErrProviderDisabled) {
		t.Fatalf("validate error = %v", err)
	}
	if _, err := adapter.VerifyWebhook(context.Background(), domain.ProviderConfiguration{}, domain.WebhookRequest{}); !errors.Is(err, ports.ErrProviderDisabled) {
		t.Fatalf("webhook error = %v", err)
	}
	if _, err := adapter.FetchConversions(context.Background(), domain.ProviderConfiguration{}, nil); !errors.Is(err, ports.ErrProviderDisabled) {
		t.Fatalf("import error = %v", err)
	}
}
