package models

import (
	"fmt"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/resiliency/health"
	"github.com/EinStack/glide/pkg/routers/latency"
	"github.com/EinStack/glide/pkg/telemetry"
)

type Config[P any] struct {
	ID          string                `yaml:"id" json:"id" validate:"required"`           // Model instance ID (unique in scope of the router)
	Enabled     bool                  `yaml:"enabled" json:"enabled" validate:"required"` // Is the model enabled?
	ErrorBudget *health.ErrorBudget   `yaml:"error_budget" json:"error_budget" swaggertype:"primitive,string"`
	Latency     *latency.Config       `yaml:"latency" json:"latency"`
	Weight      int                   `yaml:"weight" json:"weight"`
	Client      *clients.ClientConfig `yaml:"client" json:"client"`

	Provider P `yaml:"provider" json:"provider"`
}

func DefaultConfig[P any]() Config[P] {
	return Config[P]{
		Enabled:     true,
		Client:      clients.DefaultClientConfig(),
		ErrorBudget: health.DefaultErrorBudget(),
		Latency:     latency.DefaultConfig(),
		Weight:      1,
	}
}

func (c *Config) ToModel(tel *telemetry.Telemetry) (*LanguageModel, error) {
	client, err := c.Provider.ToClient(tel, c.Client)
	if err != nil {
		return nil, fmt.Errorf("error initializing client: %w", err)
	}

	return NewLangModel(c.ID, client, c.ErrorBudget, *c.Latency, c.Weight), nil
}
