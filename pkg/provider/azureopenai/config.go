package azureopenai

import (
	"github.com/EinStack/glide/pkg/config/fields"
)

// Params defines OpenAI-specific model params with the specific validation of values
// TODO: Add validations
type Params struct {
	Temperature      float64          `yaml:"temperature,omitempty" json:"temperature"`
	TopP             float64          `yaml:"top_p,omitempty" json:"top_p"`
	MaxTokens        int              `yaml:"max_tokens,omitempty" json:"max_tokens"`
	N                int              `yaml:"n,omitempty" json:"n"`
	StopWords        []string         `yaml:"stop,omitempty" json:"stop"`
	FrequencyPenalty int              `yaml:"frequency_penalty,omitempty" json:"frequency_penalty"`
	PresencePenalty  int              `yaml:"presence_penalty,omitempty" json:"presence_penalty"`
	LogitBias        *map[int]float64 `yaml:"logit_bias,omitempty" json:"logit_bias"`
	User             *string          `yaml:"user,omitempty" json:"user"`
	Seed             *int             `yaml:"seed,omitempty" json:"seed"`
	Tools            []string         `yaml:"tools,omitempty" json:"tools"`
	ToolChoice       interface{}      `yaml:"tool_choice,omitempty" json:"tool_choice"`
	ResponseFormat   interface{}      `yaml:"response_format,omitempty" json:"response_format"` // TODO: should this be a part of the chat request API?
}

func DefaultParams() Params {
	return Params{
		Temperature: 0.8,
		TopP:        1,
		MaxTokens:   100,
		N:           1,
		StopWords:   []string{},
		Tools:       []string{},
	}
}

func (p *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*p = DefaultParams()

	type plain Params // to avoid recursion

	return unmarshal((*plain)(p))
}

type Config struct {
	BaseURL       string        `yaml:"base_url" json:"base_url" validate:"required"` // The name of your Azure OpenAI Resource (e.g https://glide-test.openai.azure.com/)
	ChatEndpoint  string        `yaml:"chat_endpoint" json:"chat_endpoint"`
	ModelName     string        `yaml:"model" json:"model" validate:"required"`            // This is your deployment name. You're required to first deploy a model before you can make calls (e.g. glide-gpt-35)
	APIVersion    string        `yaml:"api_version" json:"apiVersion" validate:"required"` // The API version to use for this operation. This follows the YYYY-MM-DD format (e.g 2023-05-15)
	APIKey        fields.Secret `yaml:"api_key" json:"-" validate:"required"`
	DefaultParams *Params       `yaml:"default_params,omitempty" json:"default_params"`
}

// DefaultConfig for OpenAI models
func DefaultConfig() *Config {
	defaultParams := DefaultParams()

	return &Config{
		BaseURL:       "", // This needs to come from config
		ChatEndpoint:  "/chat/completions",
		ModelName:     "", // This needs to come from config
		APIVersion:    "2023-05-15",
		DefaultParams: &defaultParams,
	}
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = *DefaultConfig()

	type plain Config // to avoid recursion

	return unmarshal((*plain)(c))
}
