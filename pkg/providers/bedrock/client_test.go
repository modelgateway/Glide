package bedrock

import (
	"context"
	"encoding/json"
	"fmt"
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

// TODO: Need to fix this test

func TestBedrockClient_ChatRequest(t *testing.T) {
	bedrockMock := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawPayload, _ := io.ReadAll(r.Body)

		var data interface{}
		// Parse the JSON body
		err := json.Unmarshal(rawPayload, &data)
		if err != nil {
			t.Errorf("error decoding payload (%q): %v", string(rawPayload), err)
		}

		chatResponse, err := os.ReadFile(filepath.Clean("./testdata/chat.success.json"))
		if err != nil {
			t.Errorf("error reading bedrock chat mock response: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(chatResponse)
		if err != nil {
			t.Errorf("error on sending chat response: %v", err)
		}
	})

	BedrockServer := httptest.NewServer(bedrockMock)
	defer BedrockServer.Close()

	ctx := context.Background()
	providerCfg := DefaultConfig()
	clientCfg := clients.DefaultClientConfig()

	providerCfg.BaseURL = BedrockServer.URL
	providerCfg.AccessKey = "abc"
	providerCfg.SecretKey = "def"
	providerCfg.AWSRegion = "us-west-2"

	client, err := NewClient(providerCfg, clientCfg, telemetry.NewTelemetryMock())
	require.NoError(t, err)

	chatParams := schemas.ChatParams{Messages: []schemas.ChatMessage{{
		Role:    "user",
		Content: "What's the biggest animal?",
	}}}

	response, err := client.Chat(ctx, &chatParams)

	responseString := fmt.Sprintf("%+v", response)
	// errString := fmt.Sprintf("%+v", err)
	fmt.Println(responseString)

	println(response, err)
}
