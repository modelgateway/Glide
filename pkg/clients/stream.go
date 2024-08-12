package clients

import "github.com/EinStack/glide/pkg/api/schema"

type ChatStream interface {
	Open() error
	Recv() (*schema.ChatStreamChunk, error)
	Close() error
}

type ChatStreamResult struct {
	chunk *schema.ChatStreamChunk
	err   error
}

func (r *ChatStreamResult) Chunk() *schema.ChatStreamChunk {
	return r.chunk
}

func (r *ChatStreamResult) Error() error {
	return r.err
}

func NewChatStreamResult(chunk *schema.ChatStreamChunk, err error) *ChatStreamResult {
	return &ChatStreamResult{
		chunk: chunk,
		err:   err,
	}
}
