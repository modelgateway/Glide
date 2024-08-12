package retry

import (
	"time"

	"github.com/EinStack/glide/pkg/config/fields"
)

type ExpRetryConfig struct {
	MaxRetries     int              `yaml:"max_retries,omitempty" json:"max_retries"`
	BaseMultiplier int              `yaml:"base_multiplier,omitempty" json:"base_multiplier"`
	MinDelay       fields.Duration  `yaml:"min_delay,omitempty" json:"min_delay" swaggertype:"primitive,string"`
	MaxDelay       *fields.Duration `yaml:"max_delay,omitempty" json:"max_delay" swaggertype:"primitive,string"`
}

func DefaultExpRetryConfig() *ExpRetryConfig {
	maxDelay := fields.Duration(5 * time.Second)

	return &ExpRetryConfig{
		MaxRetries:     3,
		BaseMultiplier: 2,
		MinDelay:       fields.Duration(2 * time.Second),
		MaxDelay:       &maxDelay,
	}
}
