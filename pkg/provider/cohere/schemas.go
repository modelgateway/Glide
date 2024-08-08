package cohere

import "github.com/EinStack/glide/pkg/api/schemas"

// Cohere Chat Response
type ChatCompletion struct {
	Text          string                 `json:"text"`
	GenerationID  string                 `json:"generation_id"`
	ResponseID    string                 `json:"response_id"`
	TokenCount    TokenCount             `json:"token_count"`
	Citations     []Citation             `json:"citations"`
	Documents     []Documents            `json:"documents"`
	SearchQueries []SearchQuery          `json:"search_queries"`
	SearchResults []SearchResults        `json:"search_results"`
	Meta          Meta                   `json:"meta"`
	ToolInputs    map[string]interface{} `json:"tool_inputs"`
}

type TokenCount struct {
	PromptTokens   int `json:"prompt_tokens"`
	ResponseTokens int `json:"response_tokens"`
	TotalTokens    int `json:"total_tokens"`
	BilledTokens   int `json:"billed_tokens"`
}

type Meta struct {
	APIVersion struct {
		Version string `json:"version"`
	} `json:"api_version"`
	BilledUnits struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"billed_units"`
}

type Citation struct {
	Start      int      `json:"start"`
	End        int      `json:"end"`
	Text       string   `json:"text"`
	DocumentID []string `json:"document_id"`
}

type Documents struct {
	ID   string            `json:"id"`
	Data map[string]string `json:"data"` // TODO: This needs to be updated
}

type SearchQuery struct {
	Text         string `json:"text"`
	GenerationID string `json:"generation_id"`
}

type SearchResults struct {
	SearchQuery []SearchQueryObject  `json:"search_query"`
	Connectors  []ConnectorsResponse `json:"connectors"`
	DocumentID  []string             `json:"documentId"`
}

type SearchQueryObject struct {
	Text         string `json:"text"`
	GenerationID string `json:"generationId"`
}

type ConnectorsResponse struct {
	ID              string            `json:"id"`
	UserAccessToken string            `json:"user_access_token"`
	ContOnFail      string            `json:"continue_on_failure"`
	Options         map[string]string `json:"options"`
}

// ChatCompletionChunk represents SSEvent a chat response is broken down on chat streaming
// Ref: https://docs.cohere.com/reference/about
type ChatCompletionChunk struct {
	IsFinished   bool           `json:"is_finished"`
	EventType    string         `json:"event_type"`
	GenerationID *string        `json:"generation_id"`
	Text         string         `json:"text"`
	Response     *FinalResponse `json:"response,omitempty"`
	FinishReason *string        `json:"finish_reason,omitempty"`
}

type FinalResponse struct {
	ResponseID   string     `json:"response_id"`
	Text         string     `json:"text"`
	GenerationID string     `json:"generation_id"`
	TokenCount   TokenCount `json:"token_count"`
	Meta         Meta       `json:"meta"`
}

// ChatRequest is a request to complete a chat completion
// Ref: https://docs.cohere.com/reference/chat
type ChatRequest struct {
	Model             string                `json:"model"`
	Message           string                `json:"message"`
	ChatHistory       []schemas.ChatMessage `json:"chat_history"`
	Temperature       float64               `json:"temperature,omitempty"`
	Preamble          string                `json:"preamble,omitempty"`
	PromptTruncation  *string               `json:"prompt_truncation,omitempty"`
	Connectors        []string              `json:"connectors,omitempty"`
	SearchQueriesOnly bool                  `json:"search_queries_only,omitempty"`
	Stream            bool                  `json:"stream,omitempty"`
	Seed              *int                  `json:"seed,omitempty"`
	MaxTokens         *int                  `json:"max_tokens,omitempty"`
	K                 int                   `json:"k"`
	P                 float32               `json:"p"`
	FrequencyPenalty  float32               `json:"frequency_penalty"`
	PresencePenalty   float32               `json:"presence_penalty"`
	StopSequences     []string              `json:"stop_sequences"`
}

func (r *ChatRequest) ApplyParams(params *schemas.ChatParams) {
	message := params.Messages[len(params.Messages)-1]
	messageHistory := params.Messages[:len(params.Messages)-1]

	// TODO: Map chat message roles to Cohere roles: CHATBOT, SYSTEM, USER

	r.Message = message.Content
	r.ChatHistory = messageHistory
}

type Connectors struct {
	ID              string            `json:"id"`
	UserAccessToken string            `json:"user_access_token"`
	ContOnFail      string            `json:"continue_on_failure"`
	Options         map[string]string `json:"options"`
}
