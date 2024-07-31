package openai

import (
	"github.com/EinStack/glide/pkg/providers"
)

func init() {
	providers.LangRegistry.Register(ProviderOpenAI, &Config{})
}
