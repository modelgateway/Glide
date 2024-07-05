package manager

import "github.com/EinStack/glide/pkg/routers/lang"

// Config defines a config for a set of supported router types
type Config struct {
	LanguageRouters lang.RoutersConfig `yaml:"language" validate:"required,dive"` // the list of language routers
	// EmbeddingRouters []EmbeddingRouterConfig `yaml:"embedding" validate:"required,dive"`
}
