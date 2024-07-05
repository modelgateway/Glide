package config

import (
	"github.com/EinStack/glide/pkg/api"
	routerconfig "github.com/EinStack/glide/pkg/routers/manager"
	"github.com/EinStack/glide/pkg/telemetry"
)

// Config is a general top-level Glide configuration
type Config struct {
	Telemetry *telemetry.Config   `yaml:"telemetry" validate:"required"`
	API       *api.Config         `yaml:"api" validate:"required"`
	Routers   routerconfig.Config `yaml:"routers" validate:"required"`
}

func DefaultConfig() *Config {
	return &Config{
		Telemetry: telemetry.DefaultConfig(),
		API:       api.DefaultConfig(),
		// Routers should be defined by users
	}
}
