package ollama

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/clients"
)

func (c *Client) SupportChatStream() bool {
	return false
}

func (c *Client) ChatStream(_ context.Context, _ *schema.ChatParams) (clients.ChatStream, error) {
	return nil, clients.ErrChatStreamNotImplemented
}
