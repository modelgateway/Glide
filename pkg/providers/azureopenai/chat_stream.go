package azureopenai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	clients2 "github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/providers/openai"

	"github.com/r3labs/sse/v2"

	"go.uber.org/zap"

	"github.com/EinStack/glide/pkg/api/schemas"
)

// TODO: Think about reducing the number of copy-pasted code btw OpenAI and Azure providers

// ChatStream represents chat stream for a specific request
type ChatStream struct {
	tel                *telemetry.Telemetry
	client             *http.Client
	req                *http.Request
	resp               *http.Response
	reader             *sse.EventStreamReader
	finishReasonMapper *openai.FinishReasonMapper
	errMapper          *ErrorMapper
}

func NewChatStream(
	tel *telemetry.Telemetry,
	client *http.Client,
	req *http.Request,
	finishReasonMapper *openai.FinishReasonMapper,
	errMapper *ErrorMapper,
) *ChatStream {
	return &ChatStream{
		tel:                tel,
		client:             client,
		req:                req,
		finishReasonMapper: finishReasonMapper,
		errMapper:          errMapper,
	}
}

// Open initializes and opens a ChatStream.
func (s *ChatStream) Open() error {
	resp, err := s.client.Do(s.req) //nolint:bodyclose
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return s.errMapper.Map(resp)
	}

	s.resp = resp
	s.reader = sse.NewEventStreamReader(resp.Body, 4096) // TODO: should we expose maxBufferSize?

	return nil
}

// Recv receives a chat stream chunk from the ChatStream and returns a ChatStreamChunk object.
func (s *ChatStream) Recv() (*schemas.ChatStreamChunk, error) {
	var completionChunk ChatCompletionChunk

	for {
		rawEvent, err := s.reader.ReadEvent()
		if err != nil {
			s.tel.L().Warn(
				"Chat stream is unexpectedly disconnected",
				zap.String("provider", providerName),
				zap.Error(err),
			)

			// if err is io.EOF, this still means that the stream is interrupted unexpectedly
			//  because the normal stream termination is done via finding out streamDoneMarker

			return nil, clients2.ErrProviderUnavailable
		}

		s.tel.L().Debug(
			"Raw chat stream chunk",
			zap.String("provider", providerName),
			zap.ByteString("rawChunk", rawEvent),
		)

		event, err := clients2.ParseSSEvent(rawEvent)

		if bytes.Equal(event.Data, openai.StreamDoneMarker) {
			s.tel.L().Info(
				"EOF: [DONE] marker found in chat stream",
				zap.String("provider", providerName),
			)

			return nil, io.EOF
		}

		if err != nil {
			return nil, fmt.Errorf("failed to parse chat stream message: %v", err)
		}

		if !event.HasContent() {
			s.tel.L().Debug(
				"Received an empty message in chat stream, skipping it",
				zap.String("provider", providerName),
				zap.Any("msg", event),
			)

			continue
		}

		err = json.Unmarshal(event.Data, &completionChunk)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal AzureOpenAI chat stream chunk: %v", err)
		}

		responseChunk := completionChunk.Choices[0]

		// TODO: use objectpool here
		return &schemas.ChatStreamChunk{
			Cached:    false,
			Provider:  providerName,
			ModelName: completionChunk.ModelName,
			ModelResponse: schemas.ModelChunkResponse{
				Metadata: &schemas.Metadata{
					"response_id":        completionChunk.ID,
					"system_fingerprint": completionChunk.SystemFingerprint,
				},
				Message: schemas.ChatMessage{
					Role:    responseChunk.Delta.Role,
					Content: responseChunk.Delta.Content,
				},
			},
			FinishReason: s.finishReasonMapper.Map(responseChunk.FinishReason),
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

func (c *Client) ChatStream(ctx context.Context, params *schemas.ChatParams) (clients2.ChatStream, error) {
	// Create a new chat request
	httpRequest, err := c.makeStreamReq(ctx, params)
	if err != nil {
		return nil, err
	}

	return NewChatStream(
		c.tel,
		c.httpClient,
		httpRequest,
		c.finishReasonMapper,
		c.errMapper,
	), nil
}

func (c *Client) makeStreamReq(ctx context.Context, params *schemas.ChatParams) (*http.Request, error) {
	chatReq := *c.chatRequestTemplate
	chatReq.ApplyParams(params)

	chatReq.Stream = true

	rawPayload, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal AzureOpenAI chat stream request payload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatURL, bytes.NewBuffer(rawPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to create AzureOpenAI stream chat request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("api-key", string(c.config.APIKey))
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
