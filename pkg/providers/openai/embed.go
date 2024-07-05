package openai

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schemas"
)

// Embed sends an embedding request to the specified OpenAI model.
func (c *Client) Embed(_ context.Context, _ *schemas.ChatParams) (*schemas.ChatResponse, error) {
	// TODO: implement
	return nil, nil
}
