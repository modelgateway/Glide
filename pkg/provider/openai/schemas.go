package openai

import "github.com/EinStack/glide/pkg/api/schemas"

// ChatRequest is an OpenAI-specific request schema
type ChatRequest struct {
	Model            string                `json:"model"`
	Messages         []schemas.ChatMessage `json:"messages"`
	Temperature      float64               `json:"temperature,omitempty"`
	TopP             float64               `json:"top_p,omitempty"`
	MaxTokens        int                   `json:"max_tokens,omitempty"`
	N                int                   `json:"n,omitempty"`
	StopWords        []string              `json:"stop,omitempty"`
	Stream           bool                  `json:"stream,omitempty"`
	FrequencyPenalty int                   `json:"frequency_penalty,omitempty"`
	PresencePenalty  int                   `json:"presence_penalty,omitempty"`
	LogitBias        *map[int]float64      `json:"logit_bias,omitempty"`
	User             *string               `json:"user,omitempty"`
	Seed             *int                  `json:"seed,omitempty"`
	Tools            []string              `json:"tools,omitempty"`
	ToolChoice       interface{}           `json:"tool_choice,omitempty"`
	ResponseFormat   interface{}           `json:"response_format,omitempty"`
}

func (r *ChatRequest) ApplyParams(params *schemas.ChatParams) {
	// TODO(185): set other params
	r.Messages = params.Messages
}

// ChatCompletion
// Ref: https://platform.openai.com/docs/api-reference/chat/object
type ChatCompletion struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int      `json:"created"`
	ModelName         string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

type Choice struct {
	Index        int                 `json:"index"`
	Message      schemas.ChatMessage `json:"message"`
	Logprobs     interface{}         `json:"logprobs"`
	FinishReason string              `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionChunk represents SSEvent a chat response is broken down on chat streaming
// Ref: https://platform.openai.com/docs/api-reference/chat/streaming
type ChatCompletionChunk struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int            `json:"created"`
	ModelName         string         `json:"model"`
	SystemFingerprint string         `json:"system_fingerprint"`
	Choices           []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Index        int                 `json:"index"`
	Delta        schemas.ChatMessage `json:"delta"`
	Logprobs     interface{}         `json:"logprobs"`
	FinishReason string              `json:"finish_reason"`
}
