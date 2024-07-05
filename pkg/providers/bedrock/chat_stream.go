package bedrock

import (
	"context"

	clients2 "github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/api/schemas"
)

func (c *Client) SupportChatStream() bool {
	return false
}

func (c *Client) ChatStream(_ context.Context, _ *schemas.ChatParams) (clients2.ChatStream, error) {
	return nil, clients2.ErrChatStreamNotImplemented
}
