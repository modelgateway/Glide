package routers

import (
	"github.com/EinStack/glide/pkg/resiliency/retry"
	"github.com/EinStack/glide/pkg/routers/routing"
)

// TODO: how to specify other backoff strategies?
// TODO: Had to keep RoutingStrategy because of https://github.com/swaggo/swag/issues/1738

type RouterConfig struct {
	ID              string                `yaml:"id" json:"routers" validate:"required"`                                       // Unique router ID
	Enabled         bool                  `yaml:"enabled" json:"enabled" validate:"required"`                                  // Is router enabled?
	Retry           *retry.ExpRetryConfig `yaml:"retry" json:"retry" validate:"required"`                                      // retry when no healthy model is available to router
	RoutingStrategy routing.Strategy      `yaml:"strategy" json:"strategy" swaggertype:"primitive,string" validate:"required"` // strategy on picking the next model to serve the request
}

func DefaultConfig() RouterConfig {
	return RouterConfig{
		Enabled:         true,
		RoutingStrategy: routing.Priority,
		Retry:           retry.DefaultExpRetryConfig(),
	}
}
