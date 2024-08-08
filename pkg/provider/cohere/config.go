package cohere

import (
	"github.com/EinStack/glide/pkg/config/fields"
)

// Params defines Cohere-specific model params with the specific validation of values
// TODO: Add validations
type Params struct {
	Seed              *int     `yaml:"seed,omitempty" json:"seed,omitempty" validate:"omitempty,number"`
	Temperature       float64  `yaml:"temperature,omitempty" json:"temperature" validate:"required,number"`
	MaxTokens         *int     `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty" validate:"omitempty,number"`
	K                 int      `yaml:"k,omitempty" json:"k" validate:"number,gte=0,lte=500"`
	P                 float32  `yaml:"p,omitempty" json:"p" validate:"number,gte=0.01,lte=0.99"`
	FrequencyPenalty  float32  `yaml:"frequency_penalty,omitempty" json:"frequency_penalty" validate:"gte=0.0,lte=1.0"`
	PresencePenalty   float32  `yaml:"presence_penalty,omitempty" json:"presence_penalty" validate:"gte=0.0,lte=1.0"`
	Preamble          string   `yaml:"preamble,omitempty" json:"preamble,omitempty"`
	StopSequences     []string `yaml:"stop_sequences,omitempty" json:"stop_sequences" validate:"max=5"`
	PromptTruncation  *string  `yaml:"prompt_truncation,omitempty" json:"prompt_truncation,omitempty"`
	Connectors        []string `yaml:"connectors,omitempty" json:"connectors,omitempty"`
	SearchQueriesOnly bool     `yaml:"search_queries_only,omitempty" json:"search_queries_only,omitempty"`
}

func DefaultParams() Params {
	return Params{
		Temperature:       0.3,
		K:                 0,
		P:                 .75,
		SearchQueriesOnly: false,
	}
}

func (p *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*p = DefaultParams()

	type plain Params // to avoid recursion

	return unmarshal((*plain)(p))
}

type Config struct {
	BaseURL       string        `yaml:"base_url" json:"base_url" validate:"required,http_url"`
	ChatEndpoint  string        `yaml:"chat_endpoint" json:"chat_endpoint" validate:"required"`
	ModelName     string        `yaml:"model" json:"model" validate:"required"` // https://docs.cohere.com/docs/models#command
	APIKey        fields.Secret `yaml:"api_key" json:"-" validate:"required"`
	DefaultParams *Params       `yaml:"default_params,omitempty" json:"defaultParams"`
}

// DefaultConfig for Cohere models
func DefaultConfig() *Config {
	defaultParams := DefaultParams()

	return &Config{
		BaseURL:       "https://api.cohere.ai/v1",
		ChatEndpoint:  "/chat",
		ModelName:     "command-light",
		DefaultParams: &defaultParams,
	}
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = *DefaultConfig()

	type plain Config // to avoid recursion

	return unmarshal((*plain)(c))
}
