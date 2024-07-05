package azureopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/providers/openai"

	"github.com/EinStack/glide/pkg/api/schemas"

	"go.uber.org/zap"
)

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		Temperature:      cfg.DefaultParams.Temperature,
		TopP:             cfg.DefaultParams.TopP,
		MaxTokens:        cfg.DefaultParams.MaxTokens,
		N:                cfg.DefaultParams.N,
		StopWords:        cfg.DefaultParams.StopWords,
		Stream:           false,
		FrequencyPenalty: cfg.DefaultParams.FrequencyPenalty,
		PresencePenalty:  cfg.DefaultParams.PresencePenalty,
		LogitBias:        cfg.DefaultParams.LogitBias,
		User:             cfg.DefaultParams.User,
		Seed:             cfg.DefaultParams.Seed,
		Tools:            cfg.DefaultParams.Tools,
		ToolChoice:       cfg.DefaultParams.ToolChoice,
		ResponseFormat:   cfg.DefaultParams.ResponseFormat,
	}
}

// Chat sends a chat request to the specified azure openai model.
func (c *Client) Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error) {
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

func (c *Client) doChatRequest(ctx context.Context, payload *ChatRequest) (*schemas.ChatResponse, error) {
	// Build request payload
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal azure openai chat request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create azure openai chat request: %w", err)
	}

	req.Header.Set("api-key", string(c.config.APIKey))
	req.Header.Set("Content-Type", "application/json")

	// TODO: this could leak information from messages which may not be a desired thing to have
	c.tel.Logger.Debug(
		"azure openai chat request",
		zap.String("chat_url", c.chatURL),
		zap.Any("payload", payload),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send azure openai chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.errMapper.Map(resp)
	}

	// Read the response body into a byte slice
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.tel.Logger.Error("failed to read azure openai chat response", zap.Error(err))
		return nil, err
	}

	// Parse the response JSON
	var chatCompletion openai.ChatCompletion

	err = json.Unmarshal(bodyBytes, &chatCompletion)
	if err != nil {
		c.tel.Logger.Error("failed to parse openai chat response", zap.Error(err))
		return nil, err
	}

	modelChoice := chatCompletion.Choices[0]

	if len(modelChoice.Message.Content) == 0 {
		return nil, clients.ErrEmptyResponse
	}

	// Map response to UnifiedChatResponse schema
	response := schemas.ChatResponse{
		ID:        chatCompletion.ID,
		Created:   chatCompletion.Created,
		Provider:  providerName,
		ModelName: chatCompletion.ModelName,
		Cached:    false,
		ModelResponse: schemas.ModelResponse{
			Metadata: map[string]string{},
			Message: schemas.ChatMessage{
				Role:    modelChoice.Message.Role,
				Content: modelChoice.Message.Content,
			},
			TokenUsage: schemas.TokenUsage{
				PromptTokens:   chatCompletion.Usage.PromptTokens,
				ResponseTokens: chatCompletion.Usage.CompletionTokens,
				TotalTokens:    chatCompletion.Usage.TotalTokens,
			},
		},
	}

	return &response, nil
}
