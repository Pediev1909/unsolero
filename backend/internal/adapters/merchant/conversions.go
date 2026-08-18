package merchant

import (
	"context"
	"sync"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

type ConversionRegistry struct {
	mu       sync.RWMutex
	adapters map[string]ports.ConversionProviderAdapter
}

func NewConversionRegistry(adapters ...ports.ConversionProviderAdapter) *ConversionRegistry {
	registry := &ConversionRegistry{adapters: make(map[string]ports.ConversionProviderAdapter)}
	for _, adapter := range adapters {
		if adapter != nil && adapter.Key() != "" {
			registry.adapters[adapter.Key()] = adapter
		}
	}
	return registry
}

func (registry *ConversionRegistry) Select(key string) ports.ConversionProviderAdapter {
	registry.mu.RLock()
	adapter := registry.adapters[key]
	registry.mu.RUnlock()
	if adapter == nil {
		return DisabledConversionAdapter{ProviderKey: key}
	}
	return adapter
}

type DisabledConversionAdapter struct{ ProviderKey string }

func (adapter DisabledConversionAdapter) Key() string { return adapter.ProviderKey }
func (adapter DisabledConversionAdapter) ValidateConversionConfiguration(context.Context, domain.ProviderConfiguration) error {
	return ports.ErrProviderDisabled
}
func (adapter DisabledConversionAdapter) VerifyWebhook(context.Context, domain.ProviderConfiguration, domain.WebhookRequest) (domain.VerifiedWebhook, error) {
	return domain.VerifiedWebhook{}, ports.ErrProviderDisabled
}
func (adapter DisabledConversionAdapter) FetchConversions(context.Context, domain.ProviderConfiguration, *string) (domain.ConversionBatch, error) {
	return domain.ConversionBatch{}, ports.ErrProviderDisabled
}
