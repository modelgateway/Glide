package provider

import (
	"fmt"
)

var LangRegistry = NewRegistry()

type Registry struct {
	providers map[ID]Configurer
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[ID]Configurer),
	}
}

func (r *Registry) Register(name ID, config Configurer) {
	if _, ok := r.Get(name); ok {
		panic(fmt.Sprintf("provider %s is already registered", name))
	}

	r.providers[name] = config
}

func (r *Registry) Get(name ID) (Configurer, bool) {
	config, ok := r.providers[name]

	return config, ok
}

func (r *Registry) Available() []ID {
	available := make([]ID, 0, len(r.providers))

	for providerID := range r.providers {
		available = append(available, providerID)
	}

	return available
}
