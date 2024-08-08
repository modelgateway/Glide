package azureopenai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/api/schemas"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/stretchr/testify/require"
)

func TestAzureOpenAIClient_ChatRequest(t *testing.T) {
	// AzureOpenAI Chat API: https://learn.microsoft.com/en-us/azure/ai-services/openai/reference#chat-completions
	azureOpenAIMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPayload, _ := io.ReadAll(r.Body)

		var data interface{}
		// Parse the JSON body
		err := json.Unmarshal(rawPayload, &data)
		if err != nil {
			t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
		}

		chatResponse, err := os.ReadFile(filepath.Clean("./testdata/chat.success.json"))
		if err != nil {
			t.Errorf("error reading openai chat mock response: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(chatResponse)
		if err != nil {
			t.Errorf("error on sending chat response: %v", err)
		}
	})

	azureOpenAIServer := httptest.NewServer(azureOpenAIMock)
	defer azureOpenAIServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()
	providerCfg.BaseURL = azureOpenAIServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "user",
		Content: "What's the capital of the United Kingdom?",
	}}}

	response, err := client.Chat(ctx, &chatParams)
	require.NoError(t, err)

	require.Equal(t, "chatcmpl-8cdqrFT2lBQlHz0EDvvq6oQcRxNcZ", response.ID)
}

func TestAzureOpenAIClient_ChatError(t *testing.T) {
	azureOpenAIMock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	azureOpenAIServer := httptest.NewServer(azureOpenAIMock)
	defer azureOpenAIServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()
	providerCfg.BaseURL = azureOpenAIServer.URL

	// Verify the default configuration values
	require.Equal(t, "/chat/completions", providerCfg.ChatEndpoint)
	require.Equal(t, "", providerCfg.ModelName)
	require.Equal(t, "2023-05-15", providerCfg.APIVersion)
	require.NotNil(t, providerCfg.DefaultParams)

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "human",
		Content: "What's the biggest animal?",
	}}}

	response, err := client.Chat(ctx, &chatParams)
	require.Error(t, err)
	require.Nil(t, response)
}

func TestDoChatRequest_ErrorResponse(t *testing.T) {
	// Create a mock HTTP server that returns a non-OK status code
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	defer mockServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()

	providerCfg.BaseURL = mockServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "user",
		Content: "What's the dealio?",
	}}}

	_, err = client.Chat(ctx, &chatParams)

	require.Error(t, err)
	require.Contains(t, err.Error(), "provider is not available")
}
