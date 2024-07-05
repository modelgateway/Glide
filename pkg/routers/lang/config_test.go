package lang

import (
	"testing"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/providers/cohere"
	"github.com/EinStack/glide/pkg/resiliency/health"
	"github.com/EinStack/glide/pkg/resiliency/retry"
	routers2 "github.com/EinStack/glide/pkg/routers"
	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/routers/routing"

	"github.com/EinStack/glide/pkg/routers/latency"

	"github.com/EinStack/glide/pkg/providers/openai"

	"github.com/EinStack/glide/pkg/providers"

	"github.com/stretchr/testify/require"
)

func TestRouterConfig_BuildModels(t *testing.T) {
	defaultParams := openai.DefaultParams()

	cfg := routers2.Config{
		LanguageRouters: []RouterConfig{
			{
				ID:              "first_router",
				Enabled:         true,
				RoutingStrategy: routing.Priority,
				Retry:           retry.DefaultExpRetryConfig(),
				Models: []providers.LangModelConfig{
					{
						ID:          "first_model",
						Enabled:     true,
						Client:      clients.DefaultClientConfig(),
						ErrorBudget: health.DefaultErrorBudget(),
						Latency:     latency.DefaultConfig(),
						OpenAI: &openai.Config{
							APIKey:        "ABC",
							DefaultParams: &defaultParams,
						},
					},
				},
			},
			{
				ID:              "second_router",
				Enabled:         true,
				RoutingStrategy: routing.LeastLatency,
				Retry:           retry.DefaultExpRetryConfig(),
				Models: []providers.LangModelConfig{
					{
						ID:          "first_model",
						Enabled:     true,
						Client:      clients.DefaultClientConfig(),
						ErrorBudget: health.DefaultErrorBudget(),
						Latency:     latency.DefaultConfig(),
						OpenAI: &openai.Config{
							APIKey:        "ABC",
							DefaultParams: &defaultParams,
						},
					},
				},
			},
		},
	}

	routers, err := cfg.BuildLangRouters(telemetry.NewTelemetryMock())

	require.NoError(t, err)
	require.Len(t, routers, 2)
	require.Len(t, routers[0].chatModels, 1)
	require.IsType(t, &routing.PriorityRouting{}, routers[0].chatRouting)
	require.Len(t, routers[1].chatModels, 1)
	require.IsType(t, &routing.LeastLatencyRouting{}, routers[1].chatRouting)
}

func TestRouterConfig_BuildModelsPerType(t *testing.T) {
	tel := telemetry.NewTelemetryMock()
	openAIParams := openai.DefaultParams()
	cohereParams := cohere.DefaultParams()

	cfg := LangRouterConfig{
		ID:              "first_router",
		Enabled:         true,
		RoutingStrategy: routing.Priority,
		Retry:           retry.DefaultExpRetryConfig(),
		Models: []providers.LangModelConfig{
			{
				ID:          "first_model",
				Enabled:     true,
				Client:      clients.DefaultClientConfig(),
				ErrorBudget: health.DefaultErrorBudget(),
				Latency:     latency.DefaultConfig(),
				OpenAI: &openai.Config{
					APIKey:        "ABC",
					DefaultParams: &openAIParams,
				},
			},
			{
				ID:          "second_model",
				Enabled:     true,
				Client:      clients.DefaultClientConfig(),
				ErrorBudget: health.DefaultErrorBudget(),
				Latency:     latency.DefaultConfig(),
				Cohere: &cohere.Config{
					APIKey:        "ABC",
					DefaultParams: &cohereParams,
				},
			},
		},
	}

	chatModels, streamChatModels, err := cfg.BuildModels(tel)

	require.Len(t, chatModels, 2)
	require.Len(t, streamChatModels, 2)
	require.NoError(t, err)
}

func TestRouterConfig_InvalidSetups(t *testing.T) {
	defaultParams := openai.DefaultParams()

	tests := []struct {
		name   string
		config routers2.Config
	}{
		{
			"duplicated router IDs",
			routers2.Config{
				LanguageRouters: []LangRouterConfig{
					{
						ID:              "first_router",
						Enabled:         true,
						RoutingStrategy: routing.Priority,
						Retry:           retry.DefaultExpRetryConfig(),
						Models: []providers.LangModelConfig{
							{
								ID:          "first_model",
								Enabled:     true,
								Client:      clients.DefaultClientConfig(),
								ErrorBudget: health.DefaultErrorBudget(),
								Latency:     latency.DefaultConfig(),
								OpenAI: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					},
					{
						ID:              "first_router",
						Enabled:         true,
						RoutingStrategy: routing.LeastLatency,
						Retry:           retry.DefaultExpRetryConfig(),
						Models: []providers.LangModelConfig{
							{
								ID:          "first_model",
								Enabled:     true,
								Client:      clients.DefaultClientConfig(),
								ErrorBudget: health.DefaultErrorBudget(),
								Latency:     latency.DefaultConfig(),
								OpenAI: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					},
				},
			},
		},
		{
			"duplicated model IDs",
			routers2.Config{
				LanguageRouters: []LangRouterConfig{
					{
						ID:              "first_router",
						Enabled:         true,
						RoutingStrategy: routing.Priority,
						Retry:           retry.DefaultExpRetryConfig(),
						Models: []providers.LangModelConfig{
							{
								ID:          "first_model",
								Enabled:     true,
								Client:      clients.DefaultClientConfig(),
								ErrorBudget: health.DefaultErrorBudget(),
								Latency:     latency.DefaultConfig(),
								OpenAI: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
							{
								ID:          "first_model",
								Enabled:     true,
								Client:      clients.DefaultClientConfig(),
								ErrorBudget: health.DefaultErrorBudget(),
								Latency:     latency.DefaultConfig(),
								OpenAI: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					},
				},
			},
		},
		{
			"no models",
			routers2.Config{
				LanguageRouters: []LangRouterConfig{
					{
						ID:              "first_router",
						Enabled:         true,
						RoutingStrategy: routing.Priority,
						Retry:           retry.DefaultExpRetryConfig(),
						Models:          []providers.LangModelConfig{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		_, err := test.config.BuildLangRouters(telemetry.NewTelemetryMock())

		require.Error(t, err)
	}
}
