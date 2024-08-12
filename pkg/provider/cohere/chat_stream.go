package cohere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"

	"go.uber.org/zap"

	"github.com/EinStack/glide/pkg/api/schemas"
)

// SupportedEventType Cohere has other types too:
// Ref: https://docs.cohere.com/reference/chat (see Chat -> Responses -> StreamedChatResponse)
type SupportedEventType = string

var (
	StreamStartEvent SupportedEventType = "stream-start"
	TextGenEvent     SupportedEventType = "text-generation"
	StreamEndEvent   SupportedEventType = "stream-end"
)

// ChatStream represents cohere chat stream for a specific request
type ChatStream struct {
	client             *http.Client
	req                *http.Request
	modelName          string
	resp               *http.Response
	generationID       string
	streamFinished     bool
	reader             *StreamReader
	errMapper          *ErrorMapper
	finishReasonMapper *FinishReasonMapper
	tel                *telemetry.Telemetry
}

// ensure interface
var (
	_ clients.ChatStream = (*ChatStream)(nil)
)

func NewChatStream(
	tel *telemetry.Telemetry,
	client *http.Client,
	req *http.Request,
	modelName string,
	errMapper *ErrorMapper,
	finishReasonMapper *FinishReasonMapper,
) *ChatStream {
	return &ChatStream{
		tel:                tel,
		client:             client,
		req:                req,
		modelName:          modelName,
		errMapper:          errMapper,
		streamFinished:     false,
		finishReasonMapper: finishReasonMapper,
	}
}

func (s *ChatStream) Open() error {
	resp, err := s.client.Do(s.req) //nolint:bodyclose
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return s.errMapper.Map(resp)
	}

	s.tel.L().Debug("Resp Headers", zap.Any("headers", resp.Header))

	s.resp = resp
	s.reader = NewStreamReader(resp.Body, 8192) // TODO: should we expose maxBufferSize?

	return nil
}

func (s *ChatStream) Recv() (*schemas.ChatStreamChunk, error) {
	if s.streamFinished {
		return nil, io.EOF
	}

	var responseChunk ChatCompletionChunk

	for {
		rawChunk, err := s.reader.ReadEvent()
		if err != nil {
			s.tel.L().Warn(
				"Chat stream is unexpectedly disconnected",
				zap.String("provider", ProviderID),
				zap.Error(err),
			)

			// if io.EOF occurred in the middle of the stream, then the stream was interrupted

			return nil, clients.ErrProviderUnavailable
		}

		s.tel.L().Debug(
			"Raw chat stream chunk",
			zap.String("provider", ProviderID),
			zap.ByteString("rawChunk", rawChunk),
		)

		err = json.Unmarshal(rawChunk, &responseChunk)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat stream chunk: %v", err)
		}

		if responseChunk.EventType == StreamStartEvent {
			s.generationID = *responseChunk.GenerationID

			continue
		}

		if responseChunk.EventType != TextGenEvent && responseChunk.EventType != StreamEndEvent {
			s.tel.L().Debug(
				"Unsupported stream chunk type, skipping it",
				zap.String("provider", ProviderID),
				zap.ByteString("chunk", rawChunk),
			)

			continue
		}

		if responseChunk.IsFinished {
			s.streamFinished = true

			// TODO: use objectpool here
			return &schemas.ChatStreamChunk{
				Cached:    false,
				Provider:  ProviderID,
				ModelName: s.modelName,
				ModelResponse: schemas.ModelChunkResponse{
					Metadata: &schemas.Metadata{
						"generation_id": s.generationID,
						"response_id":   responseChunk.Response.ResponseID,
					},
					Message: schemas.ChatMessage{
						Role:    "model",
						Content: responseChunk.Text,
					},
				},
				FinishReason: s.finishReasonMapper.Map(responseChunk.FinishReason),
			}, nil
		}

		// TODO: use objectpool here
		return &schemas.ChatStreamChunk{
			Cached:    false,
			Provider:  ProviderID,
			ModelName: s.modelName,
			ModelResponse: schemas.ModelChunkResponse{
				Metadata: &schemas.Metadata{
					"generation_id": s.generationID,
				},
				Message: schemas.ChatMessage{
					Role:    "model",
					Content: responseChunk.Text,
				},
			},
		}, nil
	}
}

func (s *ChatStream) Close() error {
	if s.resp != nil {
		return s.resp.Body.Close()
	}

	return nil
}

func (c *Client) SupportChatStream() bool {
	return true
}

func (c *Client) ChatStream(ctx context.Context, params *schemas.ChatParams) (clients.ChatStream, error) {
	// Create a new chat request
	httpRequest, err := c.makeStreamReq(ctx, params)
	if err != nil {
		return nil, err
	}

	return NewChatStream(
		c.tel,
		c.httpClient,
		httpRequest,
		c.chatRequestTemplate.Model,
		c.errMapper,
		c.finishReasonMapper,
	), nil
}

func (c *Client) makeStreamReq(ctx context.Context, params *schemas.ChatParams) (*http.Request, error) {
	// TODO: consider using objectpool to optimize memory allocation
	chatReq := *c.chatRequestTemplate
	chatReq.ApplyParams(params)

	chatReq.Stream = true

	rawPayload, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal cohere chat stream request payload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create cohere stream chat request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", string(c.config.APIKey)))
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Accept", "text/event-stream")
	request.Header.Set("Connection", "keep-alive")

	// TODO: this could leak information from messages which may not be a desired thing to have
	c.tel.L().Debug(
		"Stream chat request",
		zap.String("chatURL", c.chatURL),
		zap.Any("payload", chatReq),
	)

	return request, nil
}
