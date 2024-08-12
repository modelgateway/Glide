package octoml

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/stretchr/testify/require"
)

func TestOctoMLClient_ChatRequest(t *testing.T) {
	// OctoML Chat API: https://docs.octoai.cloud/docs/text-gen-api-docs
	octoMLMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPayload, _ := io.ReadAll(r.Body)

		var data interface{}
		// Parse the JSON body
		err := json.Unmarshal(rawPayload, &data)
		if err != nil {
			t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
		}

		chatResponse, err := os.ReadFile(filepath.Clean("./testdata/chat.success.json"))
		if err != nil {
			t.Errorf("error reading octoml chat mock response: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(chatResponse)
		if err != nil {
			t.Errorf("error on sending chat response: %v", err)
		}
	})

	octoMLServer := httptest.NewServer(octoMLMock)
	defer octoMLServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()
	providerCfg.BaseURL = octoMLServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schema.ChatParams{Messages: []schema.ChatMessage{{
		Role:    "human",
		Content: "What's the biggest animal?",
	}}}

	response, err := client.Chat(ctx, &chatParams)
	require.NoError(t, err)

	require.Equal(t, providerCfg.ModelName, response.ModelName)
	require.Equal(t, "cmpl-8ea213aece0747aca6d0608b02b57196", response.ID)
}

func TestOctoMLClient_Chat_Error(t *testing.T) {
	// Set up the test case
	// Create a mock API server that returns an error
	octoMLMock := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Return an error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	// Create a mock API server
	octoMLServer := httptest.NewServer(octoMLMock)
	defer octoMLServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()
	providerCfg.BaseURL = octoMLServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	// Create a chat request
	chatParams := schema.ChatParams{Messages: []schema.ChatMessage{{
		Role:    "human",
		Content: "What's the biggest animal?",
	}}}

	// Call the Chat function
	_, err = client.Chat(ctx, &chatParams)

	// Check the error
	require.Error(t, err)
	require.Contains(t, err.Error(), "provider is not available")
}

func TestDoChatRequest_ErrorResponse(t *testing.T) {
	// Create a mock HTTP server that returns a non-OK status code
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	defer mockServer.Close()

	// Create a new client with the mock server URL
	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()

	providerCfg.BaseURL = mockServer.URL

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	// Create a chat request payload
	chatParams := schema.ChatParams{Messages: []schema.ChatMessage{{
		Role:    "user",
		Content: "What's the dealeo?",
	}}}

	_, err = client.Chat(ctx, &chatParams)

	require.Error(t, err)
	require.Contains(t, err.Error(), "provider is not available")
}
