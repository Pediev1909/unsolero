package ai

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"rigmark/internal/modules/ai/ports"
)

const (
	ProviderDisabled  = "disabled"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	ProviderGemini    = "gemini"
)

var (
	ErrInvalidConfiguration  = errors.New("invalid AI provider configuration")
	ErrProviderNotRegistered = errors.New("AI provider is not registered")
	providerNamePattern      = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,49}$`)
)

type Config struct {
	Provider         string
	Model            string
	APIKey           string
	Timeout          time.Duration
	MaxResponseBytes int64
}

type Factory func(Config) (ports.AIProvider, error)

type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (registry *Registry) Register(name string, factory Factory) error {
	name = strings.ToLower(strings.TrimSpace(name))
	if !providerNamePattern.MatchString(name) || name == ProviderDisabled || factory == nil {
		return fmt.Errorf("%w: provider registration", ErrInvalidConfiguration)
	}
	if _, exists := registry.factories[name]; exists {
		return fmt.Errorf("%w: duplicate provider %q", ErrInvalidConfiguration, name)
	}
	registry.factories[name] = factory
	return nil
}

// Select is the composition boundary for provider choice. OpenAI, Anthropic,
// Gemini, or another adapter registers a factory without changing application
// or domain packages.
func (registry *Registry) Select(config Config) (ports.AIProvider, error) {
	config.Provider = strings.ToLower(strings.TrimSpace(config.Provider))
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	if config.Provider == ProviderDisabled {
		return DisabledProvider{}, nil
	}
	factory, exists := registry.factories[config.Provider]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotRegistered, config.Provider)
	}
	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("create AI provider %q: %w", config.Provider, err)
	}
	if provider == nil {
		return nil, fmt.Errorf("%w: factory %q returned nil", ErrInvalidConfiguration, config.Provider)
	}
	return provider, nil
}

func validateConfig(config Config) error {
	if !providerNamePattern.MatchString(config.Provider) || config.Timeout <= 0 ||
		config.MaxResponseBytes < 1_024 || config.MaxResponseBytes > 1_048_576 {
		return fmt.Errorf("%w: provider, timeout, or response limit", ErrInvalidConfiguration)
	}
	if config.Provider != ProviderDisabled && (strings.TrimSpace(config.Model) == "" || strings.TrimSpace(config.APIKey) == "") {
		return fmt.Errorf("%w: enabled providers require a model and API key", ErrInvalidConfiguration)
	}
	return nil
}
