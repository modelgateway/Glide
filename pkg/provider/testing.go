package provider

import (
	"context"
	"io"

	"github.com/EinStack/glide/pkg/api/schema"

	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/config/fields"
	"github.com/EinStack/glide/pkg/telemetry"
)

const (
	ProviderTest = "testprovider"
)

type TestConfig struct {
	BaseURL      string        `yaml:"base_url" json:"base_url" validate:"required"`
	ChatEndpoint string        `yaml:"chat_endpoint" json:"chat_endpoint" validate:"required"`
	ModelName    string        `yaml:"model" json:"model" validate:"required"`
	APIKey       fields.Secret `yaml:"api_key" json:"-" validate:"required"`
}

func (c *TestConfig) ToClient(_ *telemetry.Telemetry, _ *clients.ClientConfig) (LangProvider, error) {
	return NewMock(nil, []RespMock{}), nil
}

func (c *TestConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain TestConfig // to avoid recursion

	return unmarshal((*plain)(c))
}

// RespMock mocks a chat response or a streaming chat chunk
type RespMock struct {
	Msg string
	Err error
}

func (m *RespMock) Resp() *schema.ChatResponse {
	return &schema.ChatResponse{
		ID: "rsp0001",
		ModelResponse: schema.ModelResponse{
			Metadata: map[string]string{
				"ID": "0001",
			},
			Message: schema.ChatMessage{
				Content: m.Msg,
			},
		},
	}
}

func (m *RespMock) RespChunk() *schema.ChatStreamChunk {
	return &schema.ChatStreamChunk{
		ModelResponse: schema.ModelChunkResponse{
			Message: schema.ChatMessage{
				Content: m.Msg,
			},
		},
	}
}

// RespStreamMock mocks a chat stream
type RespStreamMock struct {
	idx     int
	OpenErr error
	Chunks  *[]RespMock
}

// ensure interface
var (
	_ clients.ChatStream = (*RespStreamMock)(nil)
)

func NewRespStreamMock(chunk *[]RespMock) RespStreamMock {
	return RespStreamMock{
		idx:     0,
		OpenErr: nil,
		Chunks:  chunk,
	}
}

func NewRespStreamWithOpenErr(openErr error) RespStreamMock {
	return RespStreamMock{
		idx:     0,
		OpenErr: openErr,
		Chunks:  nil,
	}
}

func (m *RespStreamMock) Open() error {
	if m.OpenErr != nil {
		return m.OpenErr
	}

	return nil
}

func (m *RespStreamMock) Recv() (*schema.ChatStreamChunk, error) {
	if m.Chunks != nil && m.idx >= len(*m.Chunks) {
		return nil, io.EOF
	}

	chunks := *m.Chunks

	chunk := chunks[m.idx]
	m.idx++

	if chunk.Err != nil {
		return nil, chunk.Err
	}

	return chunk.RespChunk(), nil
}

func (m *RespStreamMock) Close() error {
	return nil
}

// Mock mocks a model provider
type Mock struct {
	idx              int
	chatResps        *[]RespMock
	chatStreams      *[]RespStreamMock
	supportStreaming bool
	modelName        *string
}

// ensure interfaces
var (
	_ LangProvider = (*Mock)(nil)
)

func NewMock(modelName *string, responses []RespMock) *Mock {
	return &Mock{
		idx:              0,
		chatResps:        &responses,
		supportStreaming: false,
		modelName:        modelName,
	}
}

func NewStreamProviderMock(modelName *string, chatStreams []RespStreamMock) *Mock {
	return &Mock{
		idx:              0,
		modelName:        modelName,
		chatStreams:      &chatStreams,
		supportStreaming: true,
	}
}

func (c *Mock) SupportChatStream() bool {
	return c.supportStreaming
}

func (c *Mock) Chat(_ context.Context, _ *schema.ChatParams) (*schema.ChatResponse, error) {
	if c.chatResps == nil {
		return nil, clients.ErrProviderUnavailable
	}

	responses := *c.chatResps

	response := responses[c.idx]
	c.idx++

	if response.Err != nil {
		return nil, response.Err
	}

	return response.Resp(), nil
}

func (c *Mock) ChatStream(_ context.Context, _ *schema.ChatParams) (clients.ChatStream, error) {
	if c.chatStreams == nil || c.idx >= len(*c.chatStreams) {
		return nil, clients.ErrProviderUnavailable
	}

	streams := *c.chatStreams

	stream := streams[c.idx]
	c.idx++

	return &stream, nil
}

func (c *Mock) Provider() string {
	return "provider_mock"
}

func (c *Mock) ModelName() string {
	if c.modelName == nil {
		return "model_mock"
	}

	return *c.modelName
}
