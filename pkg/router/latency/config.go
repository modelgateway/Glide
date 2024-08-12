package latency

import (
	"time"

	"github.com/EinStack/glide/pkg/config/fields"
)

// Config defines setting for moving average latency calculations
type Config struct {
	Decay          float64          `yaml:"decay" json:"decay"`                                                              // Weight of new latency measurements
	WarmupSamples  uint8            `yaml:"warmup_samples" json:"warmup_samples"`                                            // The number of latency probes required to init moving average
	UpdateInterval *fields.Duration `yaml:"update_interval,omitempty" json:"update_interval" swaggertype:"primitive,string"` // How often gateway should probe models with not the lowest response latency
}

func DefaultConfig() *Config {
	defaultUpdateInterval := 30 * time.Second

	return &Config{
		Decay:          0.06,
		WarmupSamples:  3,
		UpdateInterval: (*fields.Duration)(&defaultUpdateInterval),
	}
}
