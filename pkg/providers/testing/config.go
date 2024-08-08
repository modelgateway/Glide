package testing

import (
	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/config/fields"
	"github.com/EinStack/glide/pkg/provider"
	"github.com/EinStack/glide/pkg/telemetry"
)

const (
	ProviderTest = "testprovider"
)

type Config struct {
	BaseURL      string        `yaml:"base_url" json:"base_url" validate:"required"`
	ChatEndpoint string        `yaml:"chat_endpoint" json:"chat_endpoint" validate:"required"`
	ModelName    string        `yaml:"model" json:"model" validate:"required"`
	APIKey       fields.Secret `yaml:"api_key" json:"-" validate:"required"`
}

func (c *Config) ToClient(_ *telemetry.Telemetry, _ *clients.ClientConfig) (provider.LangProvider, error) {
	return NewProviderMock(nil, []RespMock{}), nil
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config // to avoid recursion

	return unmarshal((*plain)(c))
}
