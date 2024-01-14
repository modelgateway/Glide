package providers

import (
	"context"
	"errors"
	"glide/pkg/providers/clients"
	"glide/pkg/routers/health"
	"glide/pkg/routers/latency"
	"time"

	"glide/pkg/api/schemas"
)

// LangModelProvider defines an interface a provider should fulfill to be able to serve language chat requests
type LangModelProvider interface {
	Provider() string
	Chat(ctx context.Context, request *schemas.UnifiedChatRequest) (*schemas.UnifiedChatResponse, error)
}

type Model interface {
	ID() string
	Healthy() bool
}

type LanguageModel interface {
	Model
	LangModelProvider
}

// LangModel wraps provider client and expend it with health & latency tracking
type LangModel struct {
	modelID     string
	client      LangModelProvider
	rateLimit   *health.RateLimitTracker
	errorBudget *health.TokenBucket // TODO: centralize provider API health tracking in the registry
	latency     *latency.MovingAverage
}

func NewLangModel(modelID string, client LangModelProvider, budget health.ErrorBudget) *LangModel {
	return &LangModel{
		modelID:     modelID,
		client:      client,
		rateLimit:   health.NewRateLimitTracker(),
		errorBudget: health.NewTokenBucket(budget.TimePerTokenMicro(), budget.Budget()),
		latency:     latency.NewMovingAverage(0.05, 3), // TODO: set from configs
	}
}

func (m *LangModel) ID() string {
	return m.modelID
}

func (m *LangModel) Provider() string {
	return m.client.Provider()
}

func (m *LangModel) Latency() *latency.MovingAverage {
	return m.latency
}

func (m *LangModel) Healthy() bool {
	return !m.rateLimit.Limited() && m.errorBudget.HasTokens()
}

func (m *LangModel) Chat(ctx context.Context, request *schemas.UnifiedChatRequest) (*schemas.UnifiedChatResponse, error) {
	// TODO: we may want to track time-to-first-byte to "normalize" response latency wrt response size
	startedAt := time.Now()
	resp, err := m.client.Chat(ctx, request)

	// Do we want to track latency in case of errors as well?
	m.latency.Add(float64(time.Since(startedAt)))

	if err == nil {
		// successful response
		resp.ModelID = m.modelID

		return resp, err
	}

	var rle *clients.RateLimitError

	if errors.As(err, &rle) {
		m.rateLimit.SetLimited(rle.UntilReset())

		return resp, err
	}

	_ = m.errorBudget.Take(1)

	return resp, err
}
