package routers

import (
	"github.com/EinStack/glide/pkg/extmodel"
	"github.com/EinStack/glide/pkg/provider"
)

type (
	EmbedModelConfig     = extmodel.Config[*provider.Config]
	EmbedModelPoolConfig = []EmbedModelConfig
)

type EmbeddingRouterConfig struct {
	RouterConfig
	Models EmbedModelPoolConfig `yaml:"models" json:"models" validate:"required,min=1,dive"` // the list of models that could handle requests
}
