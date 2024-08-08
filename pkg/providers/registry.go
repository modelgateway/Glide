package providers

import (
	"fmt"

	"github.com/EinStack/glide/pkg/provider"
)

var LangRegistry = NewProviderRegistry()

type ProviderRegistry struct {
	providers map[provider.ProviderID]provider.ProviderConfig
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[provider.ProviderID]provider.ProviderConfig),
	}
}

func (r *ProviderRegistry) Register(name provider.ProviderID, config provider.ProviderConfig) {
	if _, ok := r.Get(name); ok {
		panic(fmt.Sprintf("provider %s is already registered", name))
	}

	r.providers[name] = config
}

func (r *ProviderRegistry) Get(name provider.ProviderID) (provider.ProviderConfig, bool) {
	config, ok := r.providers[name]

	return config, ok
}

func (r *ProviderRegistry) Available() []provider.ProviderID {
	available := make([]provider.ProviderID, 0, len(r.providers))

	for providerID := range r.providers {
		available = append(available, providerID)
	}

	return available
}
