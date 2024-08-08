package providers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/EinStack/glide/pkg/provider"
	"github.com/go-playground/validator/v10"

	"gopkg.in/yaml.v3"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/telemetry"
)

var ErrNoProviderConfigured = errors.New("exactly one provider must be configured, none is configured")

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// TODO: rename DynLangProvider to DynLangProviderConfig
type DynLangProvider map[provider.ProviderID]interface{}

var _ provider.ProviderConfig = (*DynLangProvider)(nil)

func (p DynLangProvider) ToClient(tel *telemetry.Telemetry, clientConfig *clients.ClientConfig) (provider.LangProvider, error) {
	for providerID, configValue := range p {
		if configValue == nil {
			continue
		}

		providerConfig, found := LangRegistry.Get(providerID)

		if !found {
			return nil, fmt.Errorf(
				"provider %s is not supported (available providers: %v)",
				providerID,
				strings.Join(LangRegistry.Available(), ", "),
			)
		}

		providerConfigUnmarshaller := func(providerConfig interface{}) error {
			providerConfigBytes, err := yaml.Marshal(configValue)
			if err != nil {
				return err
			}

			return yaml.Unmarshal(providerConfigBytes, providerConfig)
		}

		err := providerConfig.UnmarshalYAML(providerConfigUnmarshaller)
		if err != nil {
			return nil, err
		}

		return providerConfig.ToClient(tel, clientConfig)
	}

	return nil, provider.ErrProviderNotFound
}

// validate ensure there is only one provider configured and it's supported by Glide
func (p DynLangProvider) validate() error {
	configuredProviders := make([]provider.ProviderID, 0, len(p))

	for providerID, config := range p {
		if config != nil {
			configuredProviders = append(configuredProviders, providerID)
		}
	}

	if len(configuredProviders) == 0 {
		return ErrNoProviderConfigured
	}

	if len(configuredProviders) > 1 {
		return fmt.Errorf(
			"exactly one provider must be configured, but %v are configured (%v)",
			len(configuredProviders),
			strings.Join(configuredProviders, ", "),
		)
	}

	providerID := configuredProviders[0]
	providerConfig, found := LangRegistry.Get(providerID)

	if !found {
		return fmt.Errorf(
			"provider %s is not supported (available providers: %v)",
			providerID,
			strings.Join(LangRegistry.Available(), ", "),
		)
	}

	providerConfigUnmarshaller := func(providerConfig interface{}) error {
		configValue := p[providerID]
		providerConfigBytes, err := yaml.Marshal(configValue)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(providerConfigBytes, providerConfig)
		if err != nil {
			return err
		}

		return validate.Struct(providerConfig)
	}

	return providerConfig.UnmarshalYAML(providerConfigUnmarshaller)
}

func (p *DynLangProvider) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain DynLangProvider // to avoid recursion
	temp := plain{}

	if err := unmarshal(&temp); err != nil {
		return err
	}

	*p = DynLangProvider(temp)

	return p.validate()
}

// TODO: Remove this old LangProviders struct

//type LangProviders struct {
//	// Add other providers like
//	OpenAI      *openai.Config      `yaml:"openai,omitempty" json:"openai,omitempty"`
//	AzureOpenAI *azureopenai.Config `yaml:"azureopenai,omitempty" json:"azureopenai,omitempty"`
//	Cohere      *cohere.Config      `yaml:"cohere,omitempty" json:"cohere,omitempty"`
//	OctoML      *octoml.Config      `yaml:"octoml,omitempty" json:"octoml,omitempty"`
//	Anthropic   *anthropic.Config   `yaml:"anthropic,omitempty" json:"anthropic,omitempty"`
//	Bedrock     *bedrock.Config     `yaml:"bedrock,omitempty" json:"bedrock,omitempty"`
//	Ollama      *ollama.Config      `yaml:"ollama,omitempty" json:"ollama,omitempty"`
//}
//
//var _ ProviderConfig = (*LangProviders)(nil)

// ToClient initializes the language model client based on the provided configuration.
// It takes a telemetry object as input and returns a LangModelProvider and an error.
//func (c LangProviders) ToClient(tel *telemetry.Telemetry, clientConfig *clients.ClientConfig) (LangProvider, error) {
//	switch {
//	case c.OpenAI != nil:
//		return openai.NewClient(c.OpenAI, clientConfig, tel)
//	case c.AzureOpenAI != nil:
//		return azureopenai.NewClient(c.AzureOpenAI, clientConfig, tel)
//	case c.Cohere != nil:
//		return cohere.NewClient(c.Cohere, clientConfig, tel)
//	case c.OctoML != nil:
//		return octoml.NewClient(c.OctoML, clientConfig, tel)
//	case c.Anthropic != nil:
//		return anthropic.NewClient(c.Anthropic, clientConfig, tel)
//	case c.Bedrock != nil:
//		return bedrock.NewClient(c.Bedrock, clientConfig, tel)
//	default:
//		return nil, ErrProviderNotFound
//	}
//}

//func (c *LangProviders) validateOneProvider() error {
//	providersConfigured := 0
//
//	if c.OpenAI != nil {
//		providersConfigured++
//	}
//
//	if c.AzureOpenAI != nil {
//		providersConfigured++
//	}
//
//	if c.Cohere != nil {
//		providersConfigured++
//	}
//
//	if c.OctoML != nil {
//		providersConfigured++
//	}
//
//	if c.Anthropic != nil {
//		providersConfigured++
//	}
//
//	if c.Bedrock != nil {
//		providersConfigured++
//	}
//
//	if c.Ollama != nil {
//		providersConfigured++
//	}
//
//	// check other providers here
//	if providersConfigured == 0 {
//		return ErrNoProviderConfigured
//	}
//
//	if providersConfigured > 1 {
//		return fmt.Errorf(
//			"exactly one provider must be configured, but %v are configured",
//			providersConfigured,
//		)
//	}
//
//	return nil
//}

//func (c *LangProviders) UnmarshalYAML(unmarshal func(interface{}) error) error {
//	type plain LangProviders // to avoid recursion
//
//	if err := unmarshal((*plain)(c)); err != nil {
//		return err
//	}
//
//	return c.validateOneProvider()
//}
