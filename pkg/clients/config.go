package clients

import (
	"time"

	"github.com/EinStack/glide/pkg/config/fields"
)

type ClientConfig struct {
	Timeout             *fields.Duration `yaml:"timeout,omitempty" json:"timeout" swaggertype:"primitive,string"`
	MaxIdleConns        *int             `yaml:"max_idle_connections,omitempty" json:"max_idle_connections" swaggertype:"primitive,integer"`
	MaxIdleConnsPerHost *int             `yaml:"max_idle_connections_per_host,omitempty" json:"max_idle_connections_per_host" swaggertype:"primitive,integer"`
}

func DefaultClientConfig() *ClientConfig {
	defaultTimeout := 10 * time.Second
	maxIdleConns := 100
	maxIdleConnsPerHost := 2

	return &ClientConfig{
		Timeout:             (*fields.Duration)(&defaultTimeout),
		MaxIdleConns:        &maxIdleConns,
		MaxIdleConnsPerHost: &maxIdleConnsPerHost,
	}
}
