package providers

import (
	"fmt"
)

var LangRegistry = NewProviderRegistry()

type ProviderRegistry struct {
	providers map[ProviderID]Configurer
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[ProviderID]Configurer),
	}
}

func (r *ProviderRegistry) Register(name ProviderID, config Configurer) {
	if _, ok := r.Get(name); ok {
		panic(fmt.Sprintf("provider %s is already registered", name))
	}

	r.providers[name] = config
}

func (r *ProviderRegistry) Get(name ProviderID) (Configurer, bool) {
	config, ok := r.providers[name]

	return config, ok
}

func (r *ProviderRegistry) Available() []ProviderID {
	available := make([]ProviderID, 0, len(r.providers))

	for providerID := range r.providers {
		available = append(available, providerID)
	}

	return available
}
