package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	clients2 "github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/api/schemas"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/stretchr/testify/require"
)

func TestOpenAIClient_ChatRequest(t *testing.T) {
	// OpenAI Chat API: https://platform.openai.com/docs/api-reference/chat/create
	openAIMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	openAIServer := httptest.NewServer(openAIMock)
	defer openAIServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients2.DefaultClientConfig()

	providerCfg.BaseURL = openAIServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "user",
		Content: "What's the capital of the United Kingdom?",
	}}}

	response, err := client.Chat(ctx, &chatParams)
	require.NoError(t, err)

	require.Equal(t, "chatcmpl-123", response.ID)
}

func TestOpenAIClient_RateLimit(t *testing.T) {
	openAIMock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "5m")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	openAIServer := httptest.NewServer(openAIMock)
	defer openAIServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients2.DefaultClientConfig()

	providerCfg.BaseURL = openAIServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "human",
		Content: "What's the biggest animal?",
	}}}

	_, err = client.Chat(ctx, &chatParams)

	require.Error(t, err)
	require.IsType(t, &clients2.RateLimitError{}, err)
}
