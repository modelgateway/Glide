package config

import (
	"github.com/EinStack/glide/pkg/api"
	"github.com/EinStack/glide/pkg/routers"
	"github.com/EinStack/glide/pkg/telemetry"
)

// Config is a general top-level Glide configuration
type Config struct {
	Telemetry *telemetry.Config     `yaml:"telemetry" validate:"required"`
	API       *api.Config           `yaml:"api" validate:"required"`
	Routers   routers.RoutersConfig `yaml:"routers" validate:"required"`
}

func DefaultConfig() *Config {
	return &Config{
		Telemetry: telemetry.DefaultConfig(),
		API:       api.DefaultConfig(),
		// Routers should be defined by users
	}
}
