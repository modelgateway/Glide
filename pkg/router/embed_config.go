package router

import (
	"github.com/EinStack/glide/pkg/extmodel"
	"github.com/EinStack/glide/pkg/provider"
)

type (
	EmbedModelConfig     = extmodel.Config[*provider.Config]
	EmbedModelPoolConfig = []EmbedModelConfig
)

type EmbeddingRouterConfig struct {
	Config
	Models EmbedModelPoolConfig `yaml:"models" json:"models" validate:"required,min=1,dive"` // the list of models that could handle requests
}
