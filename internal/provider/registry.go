package provider

import (
	"fmt"
	"slices"
)

// Registry manages all available providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// A setter. Register registers a provider
func (r *Registry) Register(name string, provider Provider) {
	r.providers[name] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return provider, nil
}

// GetByModel retrieves a provider that supports the given model
func (r *Registry) GetByModel(model string) (Provider, error) {
	for _, provider := range r.providers {
		if slices.Contains(provider.GetSupportedModels(), model) {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("no provider found for model %q", model)
}

// ListModels returns all available models across all providers
func (r *Registry) ListModels() map[string][]string {
	models := make(map[string][]string)
	for name, provider := range r.providers {
		models[name] = provider.GetSupportedModels()
	}
	return models
}

// Close closes all providers
func (r *Registry) Close() error {
	var lastErr error
	for _, provider := range r.providers {
		if err := provider.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
