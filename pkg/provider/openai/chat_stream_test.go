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

func TestOpenAIClient_ChatStreamSupported(t *testing.T) {
	providerCfg := DefaultConfig()
	clientCfg := clients2.DefaultClientConfig()

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	require.True(t, client.SupportChatStream())
}

func TestOpenAIClient_ChatStreamRequest(t *testing.T) {
	tests := map[string]string{
		"success stream": "./testdata/chat_stream.success.txt",
	}

	for name, streamFile := range tests {
		t.Run(name, func(t *testing.T) {
			openAIMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rawPayload, _ := io.ReadAll(r.Body)

				var data interface{}
				// Parse the JSON body
				err := json.Unmarshal(rawPayload, &data)
				if err != nil {
					t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
				}

				chatResponse, err := os.ReadFile(filepath.Clean(streamFile))
				if err != nil {
					t.Errorf("error reading openai chat mock response: %v", err)
				}

				w.Header().Set("Content-Type", "text/event-stream")

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

			stream, err := client.ChatStream(ctx, &chatParams)
			require.NoError(t, err)

			err = stream.Open()
			require.NoError(t, err)

			for {
				chunk, err := stream.Recv()

				if err == io.EOF {
					return
				}

				require.NoError(t, err)
				require.NotNil(t, chunk)
			}
		})
	}
}

func TestOpenAIClient_ChatStreamRequestInterrupted(t *testing.T) {
	tests := map[string]string{
		"success stream, but no last done message": "./testdata/chat_stream.nodone.txt",
		"success stream, but with empty event":     "./testdata/chat_stream.empty.txt",
	}

	for name, streamFile := range tests {
		t.Run(name, func(t *testing.T) {
			openAIMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rawPayload, _ := io.ReadAll(r.Body)

				var data interface{}
				// Parse the JSON body
				err := json.Unmarshal(rawPayload, &data)
				if err != nil {
					t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
				}

				chatResponse, err := os.ReadFile(filepath.Clean(streamFile))
				if err != nil {
					t.Errorf("error reading openai chat mock response: %v", err)
				}

				w.Header().Set("Content-Type", "text/event-stream")

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

			stream, err := client.ChatStream(ctx, &chatParams)
			require.NoError(t, err)

			err = stream.Open()
			require.NoError(t, err)

			for {
				chunk, err := stream.Recv()
				if err != nil {
					require.ErrorIs(t, err, clients2.ErrProviderUnavailable)
					return
				}

				require.NoError(t, err)
				require.NotNil(t, chunk)
			}
		})
	}
}
