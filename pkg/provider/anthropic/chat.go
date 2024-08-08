package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/api/schemas"
	"go.uber.org/zap"
)

// ChatRequest is an Anthropic-specific request schema
type ChatRequest struct {
	Model         string                `json:"model"`
	Messages      []schemas.ChatMessage `json:"messages"`
	System        string                `json:"system,omitempty"`
	Temperature   float64               `json:"temperature,omitempty"`
	TopP          float64               `json:"top_p,omitempty"`
	TopK          int                   `json:"top_k,omitempty"`
	MaxTokens     int                   `json:"max_tokens,omitempty"`
	Stream        bool                  `json:"stream,omitempty"`
	Metadata      *string               `json:"metadata,omitempty"`
	StopSequences []string              `json:"stop_sequences,omitempty"`
}

func (r *ChatRequest) ApplyParams(params *schemas.ChatParams) {
	r.Messages = params.Messages
}

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		Model:         cfg.ModelName,
		System:        cfg.DefaultParams.System,
		Temperature:   cfg.DefaultParams.Temperature,
		TopP:          cfg.DefaultParams.TopP,
		TopK:          cfg.DefaultParams.TopK,
		MaxTokens:     cfg.DefaultParams.MaxTokens,
		Metadata:      cfg.DefaultParams.Metadata,
		StopSequences: cfg.DefaultParams.StopSequences,
		Stream:        false, // unsupported right now
	}
}

// Chat sends a chat request to the specified anthropic model.
//
//	Ref: https://docs.anthropic.com/claude/reference/messages_post
func (c *Client) Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error) {
	// Create a new chat request
	// TODO: consider using objectpool to optimize memory allocation
	chatReq := *c.chatRequestTemplate
	chatReq.ApplyParams(params)

	chatReq.Stream = false

	chatResponse, err := c.doChatRequest(ctx, &chatReq)
	if err != nil {
		return nil, err
	}

	return chatResponse, nil
}

func (c *Client) doChatRequest(ctx context.Context, payload *ChatRequest) (*schemas.ChatResponse, error) {
	// Build request payload
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal anthropic chat request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create anthropic chat request: %w", err)
	}

	req.Header.Set("x-api-key", string(c.config.APIKey)) // must be in lower case
	req.Header.Set("anthropic-version", c.apiVersion)
	req.Header.Set("Content-Type", "application/json")

	// TODO: this could leak information from messages which may not be a desired thing to have
	c.tel.L().Debug(
		"Anthropic chat request",
		zap.String("chat_url", c.chatURL),
		zap.Any("payload", payload),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send anthropic chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.errMapper.Map(resp)
	}

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.tel.L().Error("Failed to read anthropic chat response", zap.Error(err))
		return nil, err
	}

	// Parse the response JSON
	var anthropicResponse ChatCompletion

	err = json.Unmarshal(bodyBytes, &anthropicResponse)
	if err != nil {
		c.tel.L().Error("Failed to parse anthropic chat response", zap.Error(err))
		return nil, err
	}

	if len(anthropicResponse.Content) == 0 {
		return nil, clients.ErrEmptyResponse
	}

	completion := anthropicResponse.Content[0]

	if len(completion.Text) == 0 {
		return nil, clients.ErrEmptyResponse
	}

	usage := anthropicResponse.Usage

	// Map response to ChatResponse schema
	response := schemas.ChatResponse{
		ID:        anthropicResponse.ID,
		Created:   int(time.Now().UTC().Unix()), // not provided by anthropic
		Provider:  providerName,
		ModelName: anthropicResponse.Model,
		Cached:    false,
		ModelResponse: schemas.ModelResponse{
			Metadata: map[string]string{},
			Message: schemas.ChatMessage{
				Role:    completion.Type,
				Content: completion.Text,
			},
			TokenUsage: schemas.TokenUsage{
				PromptTokens:   usage.InputTokens,
				ResponseTokens: usage.OutputTokens,
				TotalTokens:    usage.InputTokens + usage.OutputTokens,
			},
		},
	}

	return &response, nil
}
