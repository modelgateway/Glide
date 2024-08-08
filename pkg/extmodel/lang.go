package extmodel

import (
	"context"
	"io"
	"time"

	"github.com/EinStack/glide/pkg/providers"

	"github.com/EinStack/glide/pkg/clients"
	health2 "github.com/EinStack/glide/pkg/resiliency/health"

	"github.com/EinStack/glide/pkg/config/fields"

	"github.com/EinStack/glide/pkg/routers/latency"

	"github.com/EinStack/glide/pkg/api/schemas"
)

type LangModel interface {
	Interface
	Provider() string
	ModelName() string
	Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error)
	ChatStream(ctx context.Context, params *schemas.ChatParams) (<-chan *clients.ChatStreamResult, error)
}

// LanguageModel wraps provider client and expend it with health & latency tracking
//
//	The model health is assumed to be independent of model actions (e.g. chat & chatStream)
//	The latency is assumed to be action-specific (e.g. streaming chat chunks are much low latency than the full chat action)
type LanguageModel struct {
	modelID               string
	weight                int
	client                providers.LangProvider
	healthTracker         *health2.Tracker
	chatLatency           *latency.MovingAverage
	chatStreamLatency     *latency.MovingAverage
	latencyUpdateInterval *fields.Duration
}

func NewLangModel(modelID string, client providers.LangProvider, budget *health2.ErrorBudget, latencyConfig latency.Config, weight int) *LanguageModel {
	return &LanguageModel{
		modelID:               modelID,
		client:                client,
		healthTracker:         health2.NewTracker(budget),
		chatLatency:           latency.NewMovingAverage(latencyConfig.Decay, latencyConfig.WarmupSamples),
		chatStreamLatency:     latency.NewMovingAverage(latencyConfig.Decay, latencyConfig.WarmupSamples),
		latencyUpdateInterval: latencyConfig.UpdateInterval,
		weight:                weight,
	}
}

func (m LanguageModel) ID() string {
	return m.modelID
}

func (m LanguageModel) Healthy() bool {
	return m.healthTracker.Healthy()
}

func (m LanguageModel) Weight() int {
	return m.weight
}

func (m LanguageModel) LatencyUpdateInterval() *fields.Duration {
	return m.latencyUpdateInterval
}

func (m *LanguageModel) SupportChatStream() bool {
	return m.client.SupportChatStream()
}

func (m LanguageModel) ChatLatency() *latency.MovingAverage {
	return m.chatLatency
}

func (m LanguageModel) ChatStreamLatency() *latency.MovingAverage {
	return m.chatStreamLatency
}

func (m *LanguageModel) Chat(ctx context.Context, params *schemas.ChatParams) (*schemas.ChatResponse, error) {
	startedAt := time.Now()

	resp, err := m.client.Chat(ctx, params)
	if err != nil {
		m.healthTracker.TrackErr(err)

		return resp, err
	}

	// record latency per token to normalize measurements
	m.chatLatency.Add(float64(time.Since(startedAt)) / float64(resp.ModelResponse.TokenUsage.ResponseTokens))

	// successful response
	resp.ModelID = m.modelID

	return resp, err
}

func (m *LanguageModel) ChatStream(ctx context.Context, params *schemas.ChatParams) (<-chan *clients.ChatStreamResult, error) {
	stream, err := m.client.ChatStream(ctx, params)
	if err != nil {
		m.healthTracker.TrackErr(err)

		return nil, err
	}

	startedAt := time.Now()
	err = stream.Open()
	chunkLatency := time.Since(startedAt)

	// the first chunk latency
	m.chatStreamLatency.Add(float64(chunkLatency))

	if err != nil {
		m.healthTracker.TrackErr(err)

		// if connection was not even open, we should not send our clients any messages about this failure

		return nil, err
	}

	streamResultC := make(chan *clients.ChatStreamResult)

	go func() {
		defer close(streamResultC)
		defer stream.Close()

		for {
			startedAt = time.Now()
			chunk, err := stream.Recv()
			chunkLatency = time.Since(startedAt)

			if err != nil {
				if err == io.EOF {
					// end of the stream
					return
				}

				streamResultC <- clients.NewChatStreamResult(nil, err)

				m.healthTracker.TrackErr(err)

				return
			}

			chunk.ModelID = m.modelID

			streamResultC <- clients.NewChatStreamResult(chunk, nil)

			if chunkLatency > 1*time.Millisecond {
				// All events are read in a bigger chunks of bytes, so one chunk may contain more than one event.
				//  Each byte chunk is then parsed, so there is no easy way to precisely guess latency per chunk,
				//  So we assume that if we spent more than 1ms waiting for a chunk it's likely
				//  we were trying to read from the connection (otherwise, it would take nanoseconds)
				m.chatStreamLatency.Add(float64(chunkLatency))
			}
		}
	}()

	return streamResultC, nil
}

func (m *LanguageModel) Provider() string {
	return m.client.Provider()
}

func (m *LanguageModel) ModelName() string {
	return m.client.ModelName()
}

func ChatLatency(model Interface) *latency.MovingAverage {
	return model.(LanguageModel).ChatLatency()
}

func ChatStreamLatency(model Interface) *latency.MovingAverage {
	return model.(LanguageModel).ChatStreamLatency()
}
