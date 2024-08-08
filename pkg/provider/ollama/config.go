package ollama

// Params defines Ollmama-specific model params with the specific validation of values
// TODO: Add validations
type Params struct {
	Temperature  float64  `yaml:"temperature,omitempty" json:"temperature"`
	TopP         float64  `yaml:"top_p,omitempty" json:"top_p"`
	Microstat    int      `yaml:"microstat,omitempty" json:"microstat"`
	MicrostatEta float64  `yaml:"microstat_eta,omitempty" json:"microstat_eta"`
	MicrostatTau float64  `yaml:"microstat_tau,omitempty" json:"microstat_tau"`
	NumCtx       int      `yaml:"num_ctx,omitempty" json:"num_ctx"`
	NumGqa       int      `yaml:"num_gqa,omitempty" json:"num_gqa"`
	NumGpu       int      `yaml:"num_gpu,omitempty" json:"num_gpu"`
	NumThread    int      `yaml:"num_thread,omitempty" json:"num_thread"`
	RepeatLastN  int      `yaml:"repeat_last_n,omitempty" json:"repeat_last_n"`
	Seed         int      `yaml:"seed,omitempty" json:"seed"`
	StopWords    []string `yaml:"stop,omitempty" json:"stop"`
	Tfsz         float64  `yaml:"tfs_z,omitempty" json:"tfs_z"`
	NumPredict   int      `yaml:"num_predict,omitempty" json:"num_predict"`
	TopK         int      `yaml:"top_k,omitempty" json:"top_k"`
}

func DefaultParams() Params {
	return Params{
		Temperature: 0.8,
		NumCtx:      2048,
		TopP:        0.9,
		TopK:        40,
	}
}

func (p *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*p = DefaultParams()

	type plain Params // to avoid recursion

	return unmarshal((*plain)(p))
}

type Config struct {
	BaseURL       string  `yaml:"base_url" json:"base_url" validate:"required"`
	ChatEndpoint  string  `yaml:"chat_endpoint" json:"chat_endpoint" validate:"required"`
	ModelName     string  `yaml:"model" json:"model" validate:"required"`
	DefaultParams *Params `yaml:"default_params,omitempty" json:"default_params"`
}

// DefaultConfig for OpenAI models
func DefaultConfig() *Config {
	defaultParams := DefaultParams()

	return &Config{
		BaseURL:       "http://localhost:11434",
		ChatEndpoint:  "/api/chat",
		ModelName:     "",
		DefaultParams: &defaultParams,
	}
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = *DefaultConfig()

	type plain Config // to avoid recursion

	return unmarshal((*plain)(c))
}
