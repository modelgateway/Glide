package extmodel

import (
	"fmt"

	"github.com/EinStack/glide/pkg/provider"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/resiliency/health"
	"github.com/EinStack/glide/pkg/router/latency"
	"github.com/EinStack/glide/pkg/telemetry"
)

// Config defines an extra configuration for a model wrapper around a provider
type Config[P provider.Configurer] struct {
	ID          string                `yaml:"id" json:"id" validate:"required"`           // Model instance ID (unique in scope of the router)
	Enabled     bool                  `yaml:"enabled" json:"enabled" validate:"required"` // Is the model enabled?
	ErrorBudget *health.ErrorBudget   `yaml:"error_budget" json:"error_budget" swaggertype:"primitive,string"`
	Latency     *latency.Config       `yaml:"latency" json:"latency"`
	Weight      int                   `yaml:"weight" json:"weight"`
	Client      *clients.ClientConfig `yaml:"client" json:"client"`

	Provider P `yaml:"provider" json:"provider"`
}

func NewConfig[P provider.Configurer](ID string) *Config[P] {
	config := DefaultConfig[P]()

	config.ID = ID

	return &config
}

func DefaultConfig[P provider.Configurer]() Config[P] {
	return Config[P]{
		Enabled:     true,
		Client:      clients.DefaultClientConfig(),
		ErrorBudget: health.DefaultErrorBudget(),
		Latency:     latency.DefaultConfig(),
		Weight:      1,
	}
}

func (c *Config[P]) ToModel(tel *telemetry.Telemetry) (*LanguageModel, error) {
	client, err := c.Provider.ToClient(tel, c.Client)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	return NewLangModel(c.ID, client, c.ErrorBudget, *c.Latency, c.Weight), nil
}
