package router

import (
	"testing"

	"github.com/EinStack/glide/pkg/provider"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/provider/cohere"
	"github.com/EinStack/glide/pkg/provider/openai"
	"github.com/EinStack/glide/pkg/resiliency/health"
	"github.com/EinStack/glide/pkg/router/latency"
	"github.com/EinStack/glide/pkg/router/routing"
	"github.com/EinStack/glide/pkg/telemetry"
	"github.com/stretchr/testify/require"
)

func TestRouterConfig_BuildModels(t *testing.T) {
	defaultParams := openai.DefaultParams()

	cfg := LangRoutersConfig{
		*NewRouterConfig(
			"first_router",
			WithModels(LangModelPoolConfig{
				{
					ID:          "first_model",
					Enabled:     true,
					Client:      clients.DefaultClientConfig(),
					ErrorBudget: health.DefaultErrorBudget(),
					Latency:     latency.DefaultConfig(),
					Provider: &provider.Config{
						openai.ProviderID: &openai.Config{
							APIKey:        "ABC",
							DefaultParams: &defaultParams,
						},
					},
				},
			}),
		),
		*NewRouterConfig(
			"second_router",
			WithModels(LangModelPoolConfig{
				{
					ID:          "first_model",
					Enabled:     true,
					Client:      clients.DefaultClientConfig(),
					ErrorBudget: health.DefaultErrorBudget(),
					Latency:     latency.DefaultConfig(),
					Provider: &provider.Config{
						openai.ProviderID: &openai.Config{
							APIKey:        "ABC",
							DefaultParams: &defaultParams,
						},
					},
				},
			}),
		),
	}

	routers, err := cfg.Build(telemetry.NewTelemetryMock())

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

	cfg := NewRouterConfig(
		"first_router",
		WithModels(LangModelPoolConfig{
			{
				ID:          "first_model",
				Enabled:     true,
				Client:      clients.DefaultClientConfig(),
				ErrorBudget: health.DefaultErrorBudget(),
				Latency:     latency.DefaultConfig(),
				Provider: &provider.Config{
					openai.ProviderID: &openai.Config{
						APIKey:        "ABC",
						DefaultParams: &openAIParams,
					},
				},
			},
			{
				ID:          "second_model",
				Enabled:     true,
				Client:      clients.DefaultClientConfig(),
				ErrorBudget: health.DefaultErrorBudget(),
				Latency:     latency.DefaultConfig(),
				Provider: &provider.Config{
					cohere.ProviderID: &cohere.Config{
						APIKey:        "ABC",
						DefaultParams: &cohereParams,
					},
				},
			},
		}),
	)

	chatModels, streamChatModels, err := cfg.BuildModels(tel)

	require.Len(t, chatModels, 2)
	require.Len(t, streamChatModels, 2)
	require.NoError(t, err)
}

func TestRouterConfig_InvalidSetups(t *testing.T) {
	defaultParams := openai.DefaultParams()

	tests := []struct {
		name   string
		config LangRoutersConfig
	}{
		{
			"duplicated router IDs",
			LangRoutersConfig{
				*NewRouterConfig(
					"first_router",
					WithModels(LangModelPoolConfig{
						{
							ID:          "first_model",
							Enabled:     true,
							Client:      clients.DefaultClientConfig(),
							ErrorBudget: health.DefaultErrorBudget(),
							Latency:     latency.DefaultConfig(),
							Provider: &provider.Config{
								openai.ProviderID: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					}),
				),
				*NewRouterConfig(
					"first_router",
					WithModels(LangModelPoolConfig{
						{
							ID:          "first_model",
							Enabled:     true,
							Client:      clients.DefaultClientConfig(),
							ErrorBudget: health.DefaultErrorBudget(),
							Latency:     latency.DefaultConfig(),
							Provider: &provider.Config{
								openai.ProviderID: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					}),
				),
			},
		},
		{
			"duplicated model IDs",
			LangRoutersConfig{
				*NewRouterConfig(
					"first_router",
					WithModels(LangModelPoolConfig{
						{
							ID:          "first_model",
							Enabled:     true,
							Client:      clients.DefaultClientConfig(),
							ErrorBudget: health.DefaultErrorBudget(),
							Latency:     latency.DefaultConfig(),
							Provider: &provider.Config{
								openai.ProviderID: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
						{
							ID:          "first_model",
							Enabled:     true,
							Client:      clients.DefaultClientConfig(),
							ErrorBudget: health.DefaultErrorBudget(),
							Latency:     latency.DefaultConfig(),
							Provider: &provider.Config{
								openai.ProviderID: &openai.Config{
									APIKey:        "ABC",
									DefaultParams: &defaultParams,
								},
							},
						},
					}),
				),
			},
		},
		{
			"no models",
			LangRoutersConfig{
				*NewRouterConfig(
					"first_router",
					WithModels(LangModelPoolConfig{}),
				),
			},
		},
	}

	for _, test := range tests {
		_, err := test.config.Build(telemetry.NewTelemetryMock())

		require.Error(t, err)
	}
}
