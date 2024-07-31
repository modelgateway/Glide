package provider

import (
	"context"
	"errors"

	"github.com/EinStack/glide/pkg/api/schemas"
	"github.com/EinStack/glide/pkg/clients"
)

var ErrProviderNotFound = errors.New("provider not found")

type ProviderID = string

// ModelProvider exposes provider context
type ModelProvider interface {
	Provider() ProviderID
	ModelName() string
}

// LangProvider defines an interface a provider should fulfill to be able to serve language chat requests
type LangProvider interface {
	ModelProvider

	SupportChatStream() bool

	Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error)
	ChatStream(ctx context.Context, params *schemas.ChatParams) (clients.ChatStream, error)
}

// EmbeddingProvider defines an interface a provider should fulfill to be able to generate embeddings
type EmbeddingProvider interface {
	ModelProvider

	SupportEmbedding() bool

	Embed(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error)
}
