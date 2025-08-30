package llm

import (
	"fmt"
	"strings"
)

// ProviderFactory creates LLM providers based on configuration
type ProviderFactory struct {
	configs map[string]*Config
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		configs: make(map[string]*Config),
	}
}

// RegisterProvider registers a provider configuration
func (f *ProviderFactory) RegisterProvider(name string, config *Config) {
	f.configs[strings.ToLower(name)] = config
}

// CreateProvider creates a provider instance by name
func (f *ProviderFactory) CreateProvider(name string) (Provider, error) {
	name = strings.ToLower(name)
	config, exists := f.configs[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not registered", name)
	}

	switch name {
	case "openai":
		return NewOpenAIProvider(config), nil
	case "ollama":
		return NewOllamaProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}

// GetRegisteredProviders returns a list of registered provider names
func (f *ProviderFactory) GetRegisteredProviders() []string {
	providers := make([]string, 0, len(f.configs))
	for name := range f.configs {
		providers = append(providers, name)
	}
	return providers
}
