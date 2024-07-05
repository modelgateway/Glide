package embed

import (
	"github.com/EinStack/glide/pkg/routers"
)

type EmbeddingRouterConfig struct {
	routers.RouterConfig
	// Models []providers.LangModelConfig `yaml:"models" json:"models" validate:"required,min=1,dive"` // the list of models that could handle requests
}
