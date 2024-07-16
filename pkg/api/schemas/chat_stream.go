package schemas

import "time"

type (
	Metadata     = map[string]any
	EventType    = string
	FinishReason = string
)

var (
	ReasonComplete        FinishReason = "complete"
	ReasonMaxTokens       FinishReason = "max_tokens"
	ReasonContentFiltered FinishReason = "content_filtered"
	ReasonError           FinishReason = "error"
	ReasonOther           FinishReason = "other"
)

type StreamingCacheEntry struct {
	Key            string
	Query          string
	ResponseChunks []string
	Complete       bool
}

type StreamingCacheEntryChunk struct {
	Key      string
	Index    int
	Content  ChatStreamChunk
	Complete bool
}

type StreamRequestID = string

// ChatStreamRequest defines a message that requests a new streaming chat
type ChatStreamRequest struct {
	ID StreamRequestID `json:"id" validate:"required"`
	*ChatRequest
	OverrideParams *map[string]ModelParamsOverride `json:"override_params,omitempty"`
	Metadata       *Metadata                       `json:"metadata,omitempty"`
}

func NewChatStreamFromStr(message string) *ChatStreamRequest {
	return &ChatStreamRequest{
		ChatRequest: &ChatRequest{
			Message: ChatMessage{
				"user",
				message,
			},
		},
	}
}

type ModelChunkResponse struct {
	Metadata *Metadata   `json:"metadata,omitempty"`
	Message  ChatMessage `json:"message"`
}

type ChatStreamMessage struct {
	ID        StreamRequestID  `json:"id"`
	CreatedAt int              `json:"created_at"`
	RouterID  string           `json:"router_id"`
	Metadata  *Metadata        `json:"metadata,omitempty"`
	Chunk     *ChatStreamChunk `json:"chunk,omitempty"`
	Error     *ChatStreamError `json:"error,omitempty"`
}

// ChatStreamChunk defines a message for a chunk of streaming chat response
type ChatStreamChunk struct {
	ModelID       string             `json:"model_id"`
	Provider      string             `json:"provider_id"`
	ModelName     string             `json:"model_name"`
	Cached        bool               `json:"cached"`
	ModelResponse ModelChunkResponse `json:"model_response"`
	FinishReason  *FinishReason      `json:"finish_reason,omitempty"`
}

type ChatStreamError struct {
	Name         ErrorName     `json:"name"`
	Message      string        `json:"message"`
	FinishReason *FinishReason `json:"finish_reason,omitempty"`
}

func NewChatStreamChunk(
	reqID StreamRequestID,
	routerID string,
	reqMetadata *Metadata,
	chunk *ChatStreamChunk,
) *ChatStreamMessage {
	return &ChatStreamMessage{
		ID:        reqID,
		RouterID:  routerID,
		CreatedAt: int(time.Now().UTC().Unix()),
		Metadata:  reqMetadata,
		Chunk:     chunk,
	}
}

func NewChatStreamError(
	reqID StreamRequestID,
	routerID string,
	errName ErrorName,
	errMsg string,
	reqMetadata *Metadata,
	finishReason *FinishReason,
) *ChatStreamMessage {
	return &ChatStreamMessage{
		ID:        reqID,
		RouterID:  routerID,
		CreatedAt: int(time.Now().UTC().Unix()),
		Metadata:  reqMetadata,
		Error: &ChatStreamError{
			Name:         errName,
			Message:      errMsg,
			FinishReason: finishReason,
		},
	}
}
