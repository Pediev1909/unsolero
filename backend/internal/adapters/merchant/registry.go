package merchant

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
)

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]ports.ProviderAdapter
}

func NewRegistry(adapters ...ports.ProviderAdapter) *Registry {
	registry := &Registry{adapters: make(map[string]ports.ProviderAdapter)}
	for _, adapter := range adapters {
		if adapter != nil && adapter.Key() != "" {
			registry.adapters[adapter.Key()] = adapter
		}
	}
	return registry
}

func (registry *Registry) Select(key string) ports.ProviderAdapter {
	registry.mu.RLock()
	adapter := registry.adapters[key]
	registry.mu.RUnlock()
	if adapter == nil {
		return DisabledAdapter{ProviderKey: key}
	}
	return adapter
}

type DisabledAdapter struct{ ProviderKey string }

func (adapter DisabledAdapter) Key() string { return adapter.ProviderKey }

func (adapter DisabledAdapter) ValidateConfiguration(context.Context, domain.ProviderConfiguration) error {
	return ports.ErrProviderDisabled
}

func (adapter DisabledAdapter) FetchOffers(context.Context, domain.ProviderConfiguration, *string) (domain.ProviderBatch, error) {
	return domain.ProviderBatch{}, ports.ErrProviderDisabled
}

func ProviderErrorCode(err error) string {
	switch {
	case errors.Is(err, ports.ErrProviderDisabled):
		return "provider.disabled"
	case errors.Is(err, ports.ErrProviderUnavailable):
		return "provider.unavailable"
	default:
		return "provider.failed"
	}
}

func BoundedProviderError(err error) string {
	if err == nil {
		return ""
	}
	message := fmt.Sprintf("%v", err)
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
