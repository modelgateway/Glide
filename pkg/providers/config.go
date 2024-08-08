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
