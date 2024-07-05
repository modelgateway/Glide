package models

import "github.com/EinStack/glide/pkg/config/fields"

// Model represent a configured external modality-agnostic model with its routing properties and status
type Model interface {
	ID() string
	Healthy() bool
	LatencyUpdateInterval() *fields.Duration
	Weight() int
}
