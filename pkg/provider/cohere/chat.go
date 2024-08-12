package cohere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/clients"

	"go.uber.org/zap"
)

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		Model:             cfg.ModelName,
		Temperature:       cfg.DefaultParams.Temperature,
		Preamble:          cfg.DefaultParams.Preamble,
		PromptTruncation:  cfg.DefaultParams.PromptTruncation,
		Connectors:        cfg.DefaultParams.Connectors,
		SearchQueriesOnly: cfg.DefaultParams.SearchQueriesOnly,
		Stream:            false,
	}
}

// Chat sends a chat request to the specified cohere model.
func (c *Client) Chat(ctx context.Context, params *schema.ChatParams) (*schema.ChatResponse, error) {
	// Create a new chat request
	// TODO: consider using objectpool to optimize memory allocation
	chatReq := *c.chatRequestTemplate
	chatReq.ApplyParams(params)

	chatResponse, err := c.doChatRequest(ctx, &chatReq)
	if err != nil {
		return nil, err
	}

	return chatResponse, nil
}

func (c *Client) doChatRequest(ctx context.Context, payload *ChatRequest) (*schema.ChatResponse, error) {
	// Build request payload
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal cohere chat request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create cohere chat request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+string(c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// TODO: this could leak information from messages which may not be a desired thing to have
	c.tel.Logger.Debug(
		"cohere chat request",
		zap.String("chat_url", c.chatURL),
		zap.Any("payload", payload),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send cohere chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			c.tel.Logger.Error("failed to read cohere chat response", zap.Error(err))
		}

		c.tel.Logger.Error(
			"cohere chat request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.ByteString("response", bodyBytes),
			zap.Any("headers", resp.Header),
		)

		if resp.StatusCode != http.StatusOK {
			return nil, c.errMapper.Map(resp)
		}

		// Server & client errors result in the same error to keep gateway resilient
		return nil, clients.ErrProviderUnavailable
	}

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.tel.Logger.Error("failed to read cohere chat response", zap.Error(err))
		return nil, err
	}

	// Parse the response JSON
	var cohereCompletion ChatCompletion

	err = json.Unmarshal(bodyBytes, &cohereCompletion)
	if err != nil {
		c.tel.Logger.Error("failed to parse cohere chat response", zap.Error(err))
		return nil, err
	}

	if len(cohereCompletion.Text) == 0 {
		return nil, clients.ErrEmptyResponse
	}

	// Map response to ChatResponse schema
	response := schema.ChatResponse{
		ID:        cohereCompletion.ResponseID,
		Created:   int(time.Now().UTC().Unix()), // Cohere doesn't provide this
		Provider:  ProviderID,
		ModelName: c.config.ModelName,
		Cached:    false,
		ModelResponse: schema.ModelResponse{
			Metadata: map[string]string{
				"generationId": cohereCompletion.GenerationID,
				"responseId":   cohereCompletion.ResponseID,
			},
			Message: schema.ChatMessage{
				Role:    "assistant",
				Content: cohereCompletion.Text,
			},
			TokenUsage: schema.TokenUsage{
				PromptTokens:   cohereCompletion.TokenCount.PromptTokens,
				ResponseTokens: cohereCompletion.TokenCount.ResponseTokens,
				TotalTokens:    cohereCompletion.TokenCount.TotalTokens,
			},
		},
	}

	return &response, nil
}
