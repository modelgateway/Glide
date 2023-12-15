package openai

import (
	"fmt"
	"net/http"

	"glide/pkg/providers"
)


func OpenAiApiConfig(APIKey string) pkg.ProviderApiConfig {
	return pkg.ProviderApiConfig{
		BaseURL: "https://api.openai.com/v1",
		Headers: func(APIKey string) http.Header {
			headers := make(http.Header)
			headers.Set("Authorization", fmt.Sprintf("Bearer %s", APIKey))
			headers.Set("Content-Type", "application/json")
			return headers
		},
		Complete: "/completions",
		Chat:     "/chat/completions",
		Embed:    "/embeddings",
	}
}
