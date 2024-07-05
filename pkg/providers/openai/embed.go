package openai

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schemas"
)

// Embed sends an embedding request to the specified OpenAI model.
func (c *Client) Embed(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error) {
	// TODO: implement
	return nil, nil
}
