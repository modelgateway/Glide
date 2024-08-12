package router

import (
	"fmt"

	"github.com/EinStack/glide/pkg/extmodel"
	"github.com/EinStack/glide/pkg/provider"
	"github.com/EinStack/glide/pkg/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type (
	EmbedModelConfig     = extmodel.Config[*provider.Config]
	EmbedModelPoolConfig = []EmbedModelConfig
)

type EmbedRouterConfig struct {
	Config
	Models EmbedModelPoolConfig `yaml:"models" json:"models" validate:"required,min=1,dive"` // the list of models that could handle requests
}

type EmbedRoutersConfig []EmbedRouterConfig

func (c EmbedRoutersConfig) Build(tel *telemetry.Telemetry) ([]*EmbedRouter, error) {
	seenIDs := make(map[string]bool, len(c))
	routers := make([]*EmbedRouter, 0, len(c))

	var errs error

	for idx, routerConfig := range c {
		if _, ok := seenIDs[routerConfig.ID]; ok {
			return nil, fmt.Errorf("ID \"%v\" is specified for more than one router while each ID should be unique", routerConfig.ID)
		}

		seenIDs[routerConfig.ID] = true

		if !routerConfig.Enabled {
			tel.L().Info(fmt.Sprintf("Embed router \"%v\" is disabled, skipping", routerConfig.ID))
			continue
		}

		tel.L().Debug("Init router", zap.String("routerID", routerConfig.ID))

		r, err := NewEmbedRouter(&c[idx], tel)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		routers = append(routers, r)
	}

	if errs != nil {
		return nil, errs
	}

	return routers, nil
}
