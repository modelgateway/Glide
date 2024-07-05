package cohere

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

func TestCohere_ChatStreamSupported(t *testing.T) {
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	require.True(t, client.SupportChatStream())
}

func TestCohere_ChatStreamRequest(t *testing.T) {
	tests := map[string]string{
		"success stream": "./testdata/chat_stream.success.txt",
	}

	for name, streamFile := range tests {
		t.Run(name, func(t *testing.T) {
			cohereMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rawPayload, _ := io.ReadAll(r.Body)

				var data interface{}
				// Parse the JSON body
				err := json.Unmarshal(rawPayload, &data)
				if err != nil {
					t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
				}

				chatResponse, err := os.ReadFile(filepath.Clean(streamFile))
				if err != nil {
					t.Errorf("error reading cohere chat mock response: %v", err)
				}

				w.Header().Set("Content-Type", "application/stream+json")

				_, err = w.Write(chatResponse)
				if err != nil {
					t.Errorf("error on sending chat response: %v", err)
				}
			})

			cohereServer := httptest.NewServer(cohereMock)
			defer cohereServer.Close()

			ctx := context.Background()
			providerCfg := DefaultConfig()
			clientCfg := clients.DefaultClientConfig()

			providerCfg.BaseURL = cohereServer.URL

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

func TestCohere_ChatStreamRequestInterrupted(t *testing.T) {
	tests := map[string]string{
		"interrupted stream": "./testdata/chat_stream.interrupted.txt",
	}

	for name, streamFile := range tests {
		t.Run(name, func(t *testing.T) {
			cohereMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rawPayload, _ := io.ReadAll(r.Body)

				var data interface{}
				// Parse the JSON body
				err := json.Unmarshal(rawPayload, &data)
				if err != nil {
					t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
				}

				chatResponse, err := os.ReadFile(filepath.Clean(streamFile))
				if err != nil {
					t.Errorf("error reading cohere chat mock response: %v", err)
				}

				w.Header().Set("Content-Type", "text/event-stream")

				_, err = w.Write(chatResponse)
				if err != nil {
					t.Errorf("error on sending chat response: %v", err)
				}
			})

			cohereServer := httptest.NewServer(cohereMock)
			defer cohereServer.Close()

			ctx := context.Background()
			providerCfg := DefaultConfig()
			clientCfg := clients.DefaultClientConfig()

			providerCfg.BaseURL = cohereServer.URL

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

			for range 5 {
				chunk, err := stream.Recv()

				require.NoError(t, err)
				require.NotNil(t, chunk)
			}

			chunk, err := stream.Recv()

			require.Error(t, err)
			require.Nil(t, chunk)
		})
	}
}
