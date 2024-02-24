package schemas

// ChatRequest defines Glide's Chat Request Schema unified across all language models
type ChatRequest struct {
	Message        ChatMessage         `json:"message"`
	MessageHistory []ChatMessage       `json:"messageHistory"`
	Override       OverrideChatRequest `json:"override,omitempty"`
}

type OverrideChatRequest struct {
	Model   string      `json:"model_id"`
	Message ChatMessage `json:"message"`
}

func NewChatFromStr(message string) *ChatRequest {
	return &ChatRequest{
		Message: ChatMessage{
			"human",
			message,
			"roma",
		},
	}
}

// ChatResponse defines Glide's Chat Response Schema unified across all language models
type ChatResponse struct {
	ID            string           `json:"id,omitempty"`
	Created       int              `json:"created,omitempty"`
	Provider      string           `json:"provider,omitempty"`
	RouterID      string           `json:"router,omitempty"`
	ModelID       string           `json:"model_id,omitempty"`
	Model         string           `json:"model,omitempty"`
	Cached        bool             `json:"cached,omitempty"`
	ModelResponse ProviderResponse `json:"modelResponse,omitempty"`
}

// ProviderResponse is the unified response from the provider.

type ProviderResponse struct {
	SystemID   map[string]string `json:"responseId,omitempty"`
	Message    ChatMessage       `json:"message"`
	TokenUsage TokenUsage        `json:"tokenCount"`
}

type TokenUsage struct {
	PromptTokens   float64 `json:"promptTokens"`
	ResponseTokens float64 `json:"responseTokens"`
	TotalTokens    float64 `json:"totalTokens"`
}

// ChatMessage is a message in a chat request.
type ChatMessage struct {
	// The role of the author of this message. One of system, user, or assistant.
	Role string `json:"role"`
	// The content of the message.
	Content string `json:"content"`
	// The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,
	// with a maximum length of 64 characters.
	Name string `json:"name,omitempty"`
}
