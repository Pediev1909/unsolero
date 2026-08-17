package ai

import (
	"errors"
	"strings"
	"testing"
	"time"

	"rigmark/internal/modules/ai/ports"
)

func TestRegistrySelectsConfiguredProvider(t *testing.T) {
	registry := NewRegistry()
	want := DisabledProvider{}
	if err := registry.Register("custom", func(Config) (ports.AIProvider, error) { return want, nil }); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	provider, err := registry.Select(Config{Provider: "custom", Model: "model-v1", APIKey: "secret", Timeout: time.Second, MaxResponseBytes: 4096})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if _, ok := provider.(DisabledProvider); !ok {
		t.Fatalf("Select() provider = %T, want DisabledProvider test double", provider)
	}
}

func TestRegistryRejectsUnregisteredEnabledProvider(t *testing.T) {
	_, err := NewRegistry().Select(Config{Provider: ProviderOpenAI, Model: "model-v1", APIKey: "secret", Timeout: time.Second, MaxResponseBytes: 4096})
	if !errors.Is(err, ErrProviderNotRegistered) {
		t.Fatalf("Select() error = %v, want ErrProviderNotRegistered", err)
	}
}

func TestDisabledProviderNeedsNoCredentials(t *testing.T) {
	provider, err := NewRegistry().Select(Config{Provider: ProviderDisabled, Timeout: time.Second, MaxResponseBytes: 4096})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if _, ok := provider.(DisabledProvider); !ok {
		t.Fatalf("Select() provider = %T, want DisabledProvider", provider)
	}
}

func TestDecodeStructuredJSONRejectsUnknownFieldsAndTrailingContent(t *testing.T) {
	type response struct {
		Name string `json:"name"`
	}
	for _, raw := range []string{`{"name":"known","price":42}`, `{"name":"known"} {"name":"trailing"}`} {
		if _, err := DecodeStructuredJSON[response](strings.NewReader(raw), 1024); !errors.Is(err, ErrInvalidStructuredOutput) {
			t.Errorf("DecodeStructuredJSON(%q) error = %v, want ErrInvalidStructuredOutput", raw, err)
		}
	}
}

func TestDecodeStructuredJSONRejectsOversizedResponse(t *testing.T) {
	type response struct {
		Name string `json:"name"`
	}
	if _, err := DecodeStructuredJSON[response](strings.NewReader(`{"name":"too large"}`), 5); !errors.Is(err, ErrInvalidStructuredOutput) {
		t.Fatalf("DecodeStructuredJSON() error = %v, want ErrInvalidStructuredOutput", err)
	}
}
