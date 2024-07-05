package testing

import (
	"context"
	"io"

	clients2 "github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/api/schemas"
)

// RespMock mocks a chat response or a streaming chat chunk
type RespMock struct {
	Msg string
	Err error
}

func (m *RespMock) Resp() *schemas.ChatResponse {
	return &schemas.ChatResponse{
		ID: "rsp0001",
		ModelResponse: schemas.ModelResponse{
			Metadata: map[string]string{
				"ID": "0001",
			},
			Message: schemas.ChatMessage{
				Content: m.Msg,
			},
		},
	}
}

func (m *RespMock) RespChunk() *schemas.ChatStreamChunk {
	return &schemas.ChatStreamChunk{
		ModelResponse: schemas.ModelChunkResponse{
			Message: schemas.ChatMessage{
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

func (m *RespStreamMock) Recv() (*schemas.ChatStreamChunk, error) {
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

// ProviderMock mocks a model provider
type ProviderMock struct {
	idx              int
	chatResps        *[]RespMock
	chatStreams      *[]RespStreamMock
	supportStreaming bool
	modelName        *string
}

func NewProviderMock(modelName *string, responses []RespMock) *ProviderMock {
	return &ProviderMock{
		idx:              0,
		chatResps:        &responses,
		supportStreaming: false,
		modelName:        modelName,
	}
}

func NewStreamProviderMock(modelName *string, chatStreams []RespStreamMock) *ProviderMock {
	return &ProviderMock{
		idx:              0,
		modelName:        modelName,
		chatStreams:      &chatStreams,
		supportStreaming: true,
	}
}

func (c *ProviderMock) SupportChatStream() bool {
	return c.supportStreaming
}

func (c *ProviderMock) Chat(_ context.Context, _ *schemas.ChatParams) (*schemas.ChatResponse, error) {
	if c.chatResps == nil {
		return nil, clients2.ErrProviderUnavailable
	}

	responses := *c.chatResps

	response := responses[c.idx]
	c.idx++

	if response.Err != nil {
		return nil, response.Err
	}

	return response.Resp(), nil
}

func (c *ProviderMock) ChatStream(_ context.Context, _ *schemas.ChatParams) (clients2.ChatStream, error) {
	if c.chatStreams == nil || c.idx >= len(*c.chatStreams) {
		return nil, clients2.ErrProviderUnavailable
	}

	streams := *c.chatStreams

	stream := streams[c.idx]
	c.idx++

	return &stream, nil
}

func (c *ProviderMock) Provider() string {
	return "provider_mock"
}

func (c *ProviderMock) ModelName() string {
	if c.modelName == nil {
		return "model_mock"
	}

	return *c.modelName
}
