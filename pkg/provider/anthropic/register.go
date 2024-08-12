package anthropic

import "github.com/EinStack/glide/pkg/provider"

func init() {
	provider.LangRegistry.Register(ProviderID, &Config{})
}
