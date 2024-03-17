package azureopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"glide/pkg/api/schemas"
	"glide/pkg/providers/openai"

	"go.uber.org/zap"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is an Azure openai-specific request schema
type ChatRequest struct {
	Messages         []ChatMessage    `json:"messages"`
	Temperature      float64          `json:"temperature,omitempty"`
	TopP             float64          `json:"top_p,omitempty"`
	MaxTokens        int              `json:"max_tokens,omitempty"`
	N                int              `json:"n,omitempty"`
	StopWords        []string         `json:"stop,omitempty"`
	Stream           bool             `json:"stream,omitempty"`
	FrequencyPenalty int              `json:"frequency_penalty,omitempty"`
	PresencePenalty  int              `json:"presence_penalty,omitempty"`
	LogitBias        *map[int]float64 `json:"logit_bias,omitempty"`
	User             *string          `json:"user,omitempty"`
	Seed             *int             `json:"seed,omitempty"`
	Tools            []string         `json:"tools,omitempty"`
	ToolChoice       interface{}      `json:"tool_choice,omitempty"`
	ResponseFormat   interface{}      `json:"response_format,omitempty"`
}

// NewChatRequestFromConfig fills the struct from the config. Not using reflection because of performance penalty it gives
func NewChatRequestFromConfig(cfg *Config) *ChatRequest {
	return &ChatRequest{
		Temperature:      cfg.DefaultParams.Temperature,
		TopP:             cfg.DefaultParams.TopP,
		MaxTokens:        cfg.DefaultParams.MaxTokens,
		N:                cfg.DefaultParams.N,
		StopWords:        cfg.DefaultParams.StopWords,
		Stream:           false, // unsupported right now
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

func NewChatMessagesFromUnifiedRequest(request *schemas.ChatRequest) []ChatMessage {
	messages := make([]ChatMessage, 0, len(request.MessageHistory)+1)

	// Add items from messageHistory first and the new chat message last
	for _, message := range request.MessageHistory {
		messages = append(messages, ChatMessage{Role: message.Role, Content: message.Content})
	}

	messages = append(messages, ChatMessage{Role: request.Message.Role, Content: request.Message.Content})

	return messages
}

// Chat sends a chat request to the specified azure openai model.
func (c *Client) Chat(ctx context.Context, request *schemas.ChatRequest) (*schemas.ChatResponse, error) {
	// Create a new chat request
	chatRequest := c.createChatRequestSchema(request)

	chatResponse, err := c.doChatRequest(ctx, chatRequest)
	if err != nil {
		return nil, err
	}

	if len(chatResponse.ModelResponse.Message.Content) == 0 {
		return nil, ErrEmptyResponse
	}

	return chatResponse, nil
}

func (c *Client) createChatRequestSchema(request *schemas.ChatRequest) *ChatRequest {
	// TODO: consider using objectpool to optimize memory allocation
	chatRequest := c.chatRequestTemplate // hoping to get a copy of the template
	chatRequest.Messages = NewChatMessagesFromUnifiedRequest(request)

	return chatRequest
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
	c.telemetry.Logger.Debug(
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
		c.telemetry.Logger.Error("failed to read azure openai chat response", zap.Error(err))
		return nil, err
	}

	// Parse the response JSON
	var openAICompletion openai.ChatCompletion

	err = json.Unmarshal(bodyBytes, &openAICompletion)
	if err != nil {
		c.telemetry.Logger.Error("failed to parse openai chat response", zap.Error(err))
		return nil, err
	}

	openAICompletion.SystemFingerprint = "" // Azure OpenAI doesn't return this

	// Map response to UnifiedChatResponse schema
	response := schemas.ChatResponse{
		ID:        openAICompletion.ID,
		Created:   openAICompletion.Created,
		Provider:  providerName,
		ModelName: openAICompletion.ModelName,
		Cached:    false,
		ModelResponse: schemas.ModelResponse{
			SystemID: map[string]string{
				"system_fingerprint": openAICompletion.SystemFingerprint,
			},
			Message: schemas.ChatMessage{
				Role:    openAICompletion.Choices[0].Message.Role,
				Content: openAICompletion.Choices[0].Message.Content,
			},
			TokenUsage: schemas.TokenUsage{
				PromptTokens:   openAICompletion.Usage.PromptTokens,
				ResponseTokens: openAICompletion.Usage.CompletionTokens,
				TotalTokens:    openAICompletion.Usage.TotalTokens,
			},
		},
	}

	return &response, nil
}
