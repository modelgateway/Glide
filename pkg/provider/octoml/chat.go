package octoml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/provider/openai"

	"go.uber.org/zap"
)

// ChatRequest is an octoml-specific request schema
type ChatRequest struct {
	Model            string               `json:"model"`
	Messages         []schema.ChatMessage `json:"messages"`
	Temperature      float64              `json:"temperature,omitempty"`
	TopP             float64              `json:"top_p,omitempty"`
	MaxTokens        int                  `json:"max_tokens,omitempty"`
	StopWords        []string             `json:"stop,omitempty"`
	Stream           bool                 `json:"stream,omitempty"`
	FrequencyPenalty int                  `json:"frequency_penalty,omitempty"`
	PresencePenalty  int                  `json:"presence_penalty,omitempty"`
}

func (r *ChatRequest) ApplyParams(params *schema.ChatParams) {
	// TODO(185): set other params
	r.Messages = params.Messages
}

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		Model:            cfg.ModelName,
		Temperature:      cfg.DefaultParams.Temperature,
		TopP:             cfg.DefaultParams.TopP,
		MaxTokens:        cfg.DefaultParams.MaxTokens,
		StopWords:        cfg.DefaultParams.StopWords,
		FrequencyPenalty: cfg.DefaultParams.FrequencyPenalty,
		PresencePenalty:  cfg.DefaultParams.PresencePenalty,
	}
}

// Chat sends a chat request to the specified octoml model.
func (c *Client) Chat(ctx context.Context, params *schema.ChatParams) (*schema.ChatResponse, error) {
	// Create a new chat request
	// TODO: consider using objectpool to optimize memory allocation
	chatReq := *c.chatRequestTemplate // hoping to get a copy of the template
	chatReq.ApplyParams(params)

	chatReq.Stream = false

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
		return nil, fmt.Errorf("unable to marshal octoml chat request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create octoml chat request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+string(c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// TODO: this could leak information from messages which may not be a desired thing to have
	c.telemetry.Logger.Debug(
		"octoml chat request",
		zap.String("chat_url", c.chatURL),
		zap.Any("payload", payload),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send octoml chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.errMapper.Map(resp)
	}

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.telemetry.Logger.Error("failed to read octoml chat response", zap.Error(err))
		return nil, err
	}

	// Parse the response JSON
	var completion openai.ChatCompletion // Octo uses the same response schema as OpenAI

	err = json.Unmarshal(bodyBytes, &completion)
	if err != nil {
		c.telemetry.Logger.Error("failed to parse openai chat response", zap.Error(err))
		return nil, err
	}

	modelChoice := completion.Choices[0]

	if len(modelChoice.Message.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	// Map response to UnifiedChatResponse schema
	response := schema.ChatResponse{
		ID:        completion.ID,
		Created:   completion.Created,
		Provider:  providerName,
		ModelName: completion.ModelName,
		Cached:    false,
		ModelResponse: schema.ModelResponse{
			Metadata: map[string]string{
				"system_fingerprint": completion.SystemFingerprint,
			},
			Message: schema.ChatMessage{
				Role:    modelChoice.Message.Role,
				Content: modelChoice.Message.Content,
			},
			TokenUsage: schema.TokenUsage{
				PromptTokens:   completion.Usage.PromptTokens,
				ResponseTokens: completion.Usage.CompletionTokens,
				TotalTokens:    completion.Usage.TotalTokens,
			},
		},
	}

	return &response, nil
}
