package extmodel

import "github.com/EinStack/glide/pkg/config/fields"

// Interface represent a configured external modality-agnostic model with its routing properties and status
type Interface interface {
	ID() string
	Healthy() bool
	LatencyUpdateInterval() *fields.Duration
	Weight() int
}
