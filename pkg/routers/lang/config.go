package lang

import (
	"fmt"
	"github.com/EinStack/glide/pkg/providers"
	"time"

	"github.com/EinStack/glide/pkg/routers"

	"github.com/EinStack/glide/pkg/models"
	"github.com/EinStack/glide/pkg/resiliency/retry"
	"github.com/EinStack/glide/pkg/routers/routing"
	"github.com/EinStack/glide/pkg/telemetry"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type (
	ModelConfig     = models.Config[providers.DynLangProvider]
	ModelPoolConfig = []ModelConfig
)

// RouterConfig
type RouterConfig struct {
	routers.RouterConfig
	Models ModelPoolConfig `yaml:"models" json:"models" validate:"required,min=1,dive"` // the list of models that could handle requests
}

type RouterConfigOption = func(*RouterConfig)

func WithModels(models ModelPoolConfig) RouterConfigOption {
	return func(c *RouterConfig) {
		c.Models = models
	}
}

func NewRouterConfig(RouterID string, opt ...RouterConfigOption) *RouterConfig {
	config := &RouterConfig{
		RouterConfig: routers.DefaultConfig(),
	}

	config.ID = RouterID

	for _, o := range opt {
		o(config)
	}

	return config
}

// BuildModels creates LanguageModel slice out of the given config
func (c *RouterConfig) BuildModels(tel *telemetry.Telemetry) ([]*models.LanguageModel, []*models.LanguageModel, error) { //nolint: cyclop
	var errs error

	seenIDs := make(map[string]bool, len(c.Models))
	chatModels := make([]*models.LanguageModel, 0, len(c.Models))
	chatStreamModels := make([]*models.LanguageModel, 0, len(c.Models))

	for _, modelConfig := range c.Models {
		if _, ok := seenIDs[modelConfig.ID]; ok {
			return nil, nil, fmt.Errorf(
				"ID \"%v\" is specified for more than one model in router \"%v\", while it should be unique in scope of that pool",
				modelConfig.ID,
				c.ID,
			)
		}

		seenIDs[modelConfig.ID] = true

		if !modelConfig.Enabled {
			tel.L().Info(
				"ModelName is disabled, skipping",
				zap.String("router", c.ID),
				zap.String("model", modelConfig.ID),
			)

			continue
		}

		tel.L().Debug(
			"Init lang model",
			zap.String("router", c.ID),
			zap.String("model", modelConfig.ID),
		)

		model, err := modelConfig.ToModel(tel)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		chatModels = append(chatModels, model)

		if !model.SupportChatStream() {
			tel.L().WithOptions(zap.AddStacktrace(zap.ErrorLevel)).Warn(
				"Provider doesn't support or have not been yet integrated with streaming chat, it won't serve streaming chat requests",
				zap.String("routerID", c.ID),
				zap.String("modelID", model.ID()),
				zap.String("provider", model.Provider()),
			)

			continue
		}

		chatStreamModels = append(chatStreamModels, model)
	}

	if errs != nil {
		return nil, nil, errs
	}

	if len(chatModels) == 0 {
		return nil, nil, fmt.Errorf("router \"%v\" must have at least one active model, zero defined", c.ID)
	}

	if len(chatModels) == 1 {
		tel.L().WithOptions(zap.AddStacktrace(zap.ErrorLevel)).Warn(
			fmt.Sprintf("Router \"%v\" has only one active model defined. "+
				"This is not recommended for production setups. "+
				"Define at least a few models to leverage resiliency logic Glide provides",
				c.ID,
			),
		)
	}

	if len(chatStreamModels) == 1 {
		tel.L().WithOptions(zap.AddStacktrace(zap.ErrorLevel)).Warn(
			fmt.Sprintf("Router \"%v\" has only one active model defined with streaming chat support. "+
				"This is not recommended for production setups. "+
				"Define at least a few models to leverage resiliency logic Glide provides",
				c.ID,
			),
		)
	}

	if len(chatStreamModels) == 0 {
		tel.L().WithOptions(zap.AddStacktrace(zap.ErrorLevel)).Warn(
			fmt.Sprintf("Router \"%v\" has only no model with streaming chat support. "+
				"The streaming chat workflow won't work until you define any",
				c.ID,
			),
		)
	}

	return chatModels, chatStreamModels, nil
}

func (c *RouterConfig) BuildRetry() *retry.ExpRetry {
	retryConfig := c.Retry
	maxDelay := time.Duration(*retryConfig.MaxDelay)

	return retry.NewExpRetry(
		retryConfig.MaxRetries,
		retryConfig.BaseMultiplier,
		time.Duration(retryConfig.MinDelay),
		&maxDelay,
	)
}

func (c *RouterConfig) BuildRouting(
	chatModels []*models.LanguageModel,
	chatStreamModels []*models.LanguageModel,
) (routing.LangModelRouting, routing.LangModelRouting, error) {
	chatModelPool := make([]models.Model, 0, len(chatModels))
	chatStreamModelPool := make([]models.Model, 0, len(chatStreamModels))

	for _, model := range chatModels {
		chatModelPool = append(chatModelPool, model)
	}

	for _, model := range chatStreamModels {
		chatStreamModelPool = append(chatStreamModelPool, model)
	}

	switch c.RoutingStrategy {
	case routing.Priority:
		return routing.NewPriority(chatModelPool), routing.NewPriority(chatStreamModelPool), nil
	case routing.RoundRobin:
		return routing.NewRoundRobinRouting(chatModelPool), routing.NewRoundRobinRouting(chatStreamModelPool), nil
	case routing.WeightedRoundRobin:
		return routing.NewWeightedRoundRobin(chatModelPool), routing.NewWeightedRoundRobin(chatStreamModelPool), nil
	case routing.LeastLatency:
		return routing.NewLeastLatencyRouting(models.ChatLatency, chatModelPool),
			routing.NewLeastLatencyRouting(models.ChatStreamLatency, chatStreamModelPool),
			nil
	}

	return nil, nil, fmt.Errorf("routing strategy \"%v\" is not supported, please make sure there is no typo", c.RoutingStrategy)
}

func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		RouterConfig: routers.DefaultConfig(),
	}
}

func (c *RouterConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = *DefaultRouterConfig()

	type plain RouterConfig // to avoid recursion

	return unmarshal((*plain)(c))
}

type RoutersConfig []RouterConfig

func (c RoutersConfig) Build(tel *telemetry.Telemetry) ([]*Router, error) {
	seenIDs := make(map[string]bool, len(c))
	langRouters := make([]*Router, 0, len(c))

	var errs error

	for idx, routerConfig := range c {
		if _, ok := seenIDs[routerConfig.ID]; ok {
			return nil, fmt.Errorf("ID \"%v\" is specified for more than one router while each ID should be unique", routerConfig.ID)
		}

		seenIDs[routerConfig.ID] = true

		if !routerConfig.Enabled {
			tel.L().Info(fmt.Sprintf("Router \"%v\" is disabled, skipping", routerConfig.ID))
			continue
		}

		tel.L().Debug("Init router", zap.String("routerID", routerConfig.ID))

		router, err := NewLangRouter(&c[idx], tel)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}

		langRouters = append(langRouters, router)
	}

	if errs != nil {
		return nil, errs
	}

	return langRouters, nil
}
