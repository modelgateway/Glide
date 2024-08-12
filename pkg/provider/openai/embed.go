package openai

import (
	"context"

	"github.com/EinStack/glide/pkg/api/schema"
)

// Embed sends an embedding request to the specified OpenAI model.
func (c *Client) Embed(_ context.Context, _ *schema.ChatParams) (*schema.ChatResponse, error) {
	// TODO: implement
	return nil, nil
}
