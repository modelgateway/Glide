package routers

import (
	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/telemetry"
)

type RouterManager struct {
	Config        *RoutersConfig
	tel           *telemetry.Telemetry
	langRouterMap *map[string]*LangRouter
	langRouters   []*LangRouter
}

// NewManager creates a new instance of Router Manager that creates, holds and returns all routers
func NewManager(cfg *RoutersConfig, tel *telemetry.Telemetry) (*RouterManager, error) {
	langRouters, err := cfg.LanguageRouters.Build(tel)
	if err != nil {
		return nil, err
	}

	langRouterMap := make(map[string]*LangRouter, len(langRouters))

	for _, router := range langRouters {
		langRouterMap[router.ID()] = router
	}

	manager := RouterManager{
		Config:        cfg,
		tel:           tel,
		langRouters:   langRouters,
		langRouterMap: &langRouterMap,
	}

	return &manager, err
}

func (r *RouterManager) GetLangRouters() []*LangRouter {
	return r.langRouters
}

// GetLangRouter returns a router by type and ID
func (r *RouterManager) GetLangRouter(routerID string) (*LangRouter, error) {
	if router, found := (*r.langRouterMap)[routerID]; found {
		return router, nil
	}

	return nil, &schemas.ErrRouterNotFound
}
