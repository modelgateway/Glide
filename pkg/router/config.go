package router

import (
	"github.com/EinStack/glide/pkg/resiliency/retry"
	"github.com/EinStack/glide/pkg/router/routing"
)

// TODO: how to specify other backoff strategies?
// TODO: Had to keep RoutingStrategy because of https://github.com/swaggo/swag/issues/1738

type Config struct {
	ID              string                `yaml:"id" json:"routers" validate:"required"`                                       // Unique router ID
	Enabled         bool                  `yaml:"enabled" json:"enabled" validate:"required"`                                  // Is router enabled?
	Retry           *retry.ExpRetryConfig `yaml:"retry" json:"retry" validate:"required"`                                      // retry when no healthy model is available to router
	RoutingStrategy routing.Strategy      `yaml:"strategy" json:"strategy" swaggertype:"primitive,string" validate:"required"` // strategy on picking the next model to serve the request
}

func DefaultConfig() Config {
	return Config{
		Enabled:         true,
		RoutingStrategy: routing.Priority,
		Retry:           retry.DefaultExpRetryConfig(),
	}
}

// RoutersConfig defines a config for a set of supported router types
type RoutersConfig struct {
	LanguageRouters  LangRoutersConfig  `yaml:"language" validate:"required,dive"` // the list of language routers
	EmbeddingRouters EmbedRoutersConfig `yaml:"embedding" validate:"required,dive"`
}
