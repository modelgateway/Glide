package providers

import (
	"context"
	"errors"

	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/clients"
)

var ErrProviderNotFound = errors.New("provider not found")

// ModelProvider exposes provider context
type ModelProvider interface {
	Provider() string
	ModelName() string
}

// LangProvider defines an interface a provider should fulfill to be able to serve language chat requests
type LangProvider interface {
	ModelProvider

	SupportChatStream() bool

	Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error)
	ChatStream(ctx context.Context, params *schemas.ChatParams) (clients.ChatStream, error)
}
