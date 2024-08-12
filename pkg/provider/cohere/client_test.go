// pkg/providers/cohere/client_test.go
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

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/stretchr/testify/require"
)

func TestCohereClient_ChatRequest(t *testing.T) {
	cohereMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPayload, _ := io.ReadAll(r.Body)

		var data interface{}
		// Parse the JSON body
		err := json.Unmarshal(rawPayload, &data)
		if err != nil {
			t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
		}

		chatResponse, err := os.ReadFile(filepath.Clean("./testdata/chat.success.json"))
		if err != nil {
			t.Errorf("error reading cohere chat mock response: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")

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

	chatParams := schema.ChatParams{Messages: []schema.ChatMessage{{
		Role:    "human",
		Content: "What's the biggest animal?",
	}}}

	response, err := client.Chat(ctx, &chatParams)
	require.NoError(t, err)

	require.Equal(t, "ec9eb88b-2da5-462e-8f0f-0899d243aa2e", response.ID)
}
