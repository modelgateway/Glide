package providers

import (
	"errors"
	"fmt"

	"github.com/EinStack/glide/pkg/providers/ollama"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/providers/anthropic"
	"github.com/EinStack/glide/pkg/providers/azureopenai"
	"github.com/EinStack/glide/pkg/providers/bedrock"
	"github.com/EinStack/glide/pkg/providers/cohere"
	"github.com/EinStack/glide/pkg/providers/octoml"
	"github.com/EinStack/glide/pkg/providers/openai"
	"github.com/EinStack/glide/pkg/telemetry"
)

var ErrProviderNotFound = errors.New("provider not found")

type LangProviders struct {
	// Add other providers like
	OpenAI      *openai.Config      `yaml:"openai,omitempty" json:"openai,omitempty"`
	AzureOpenAI *azureopenai.Config `yaml:"azureopenai,omitempty" json:"azureopenai,omitempty"`
	Cohere      *cohere.Config      `yaml:"cohere,omitempty" json:"cohere,omitempty"`
	OctoML      *octoml.Config      `yaml:"octoml,omitempty" json:"octoml,omitempty"`
	Anthropic   *anthropic.Config   `yaml:"anthropic,omitempty" json:"anthropic,omitempty"`
	Bedrock     *bedrock.Config     `yaml:"bedrock,omitempty" json:"bedrock,omitempty"`
	Ollama      *ollama.Config      `yaml:"ollama,omitempty" json:"ollama,omitempty"`
}

// ToClient initializes the language model client based on the provided configuration.
// It takes a telemetry object as input and returns a LangModelProvider and an error.
func (c *LangProviders) ToClient(tel *telemetry.Telemetry, clientConfig *clients.ClientConfig) (LangProvider, error) {
	switch {
	case c.OpenAI != nil:
		return openai.NewClient(c.OpenAI, clientConfig, tel)
	case c.AzureOpenAI != nil:
		return azureopenai.NewClient(c.AzureOpenAI, clientConfig, tel)
	case c.Cohere != nil:
		return cohere.NewClient(c.Cohere, clientConfig, tel)
	case c.OctoML != nil:
		return octoml.NewClient(c.OctoML, clientConfig, tel)
	case c.Anthropic != nil:
		return anthropic.NewClient(c.Anthropic, clientConfig, tel)
	case c.Bedrock != nil:
		return bedrock.NewClient(c.Bedrock, clientConfig, tel)
	default:
		return nil, ErrProviderNotFound
	}
}

func (c *LangProviders) validateOneProvider() error {
	providersConfigured := 0

	if c.OpenAI != nil {
		providersConfigured++
	}

	if c.AzureOpenAI != nil {
		providersConfigured++
	}

	if c.Cohere != nil {
		providersConfigured++
	}

	if c.OctoML != nil {
		providersConfigured++
	}

	if c.Anthropic != nil {
		providersConfigured++
	}

	if c.Bedrock != nil {
		providersConfigured++
	}

	if c.Ollama != nil {
		providersConfigured++
	}

	// check other providers here
	if providersConfigured == 0 {
		return fmt.Errorf("exactly one provider must be configured, none is configured")
	}

	if providersConfigured > 1 {
		return fmt.Errorf(
			"exactly one provider must be configured, but %v are configured",
			providersConfigured,
		)
	}

	return nil
}

func (c *LangProviders) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig()

	type plain LangModelConfig // to avoid recursion

	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return c.validateOneProvider()
}
