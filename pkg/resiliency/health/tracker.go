package health

import (
	"errors"

	"github.com/EinStack/glide/pkg/clients"
)

// Tracker tracks errors and general health of model provider
type Tracker struct {
	unauthorized bool
	errBudget    *TokenBucket
	rateLimit    *RateLimitTracker
}

func NewTracker(budget *ErrorBudget) *Tracker {
	return &Tracker{
		unauthorized: false,
		rateLimit:    NewRateLimitTracker(),
		errBudget:    NewTokenBucket(budget.TimePerTokenMicro(), budget.Budget()),
	}
}

func (t *Tracker) Healthy() bool {
	return !t.unauthorized && !t.rateLimit.Limited() && t.errBudget.HasTokens()
}

func (t *Tracker) TrackErr(err error) {
	var rateLimitErr *clients.RateLimitError

	if errors.Is(err, clients.ErrUnauthorized) {
		t.unauthorized = true

		return
	}

	if errors.As(err, &rateLimitErr) {
		t.rateLimit.SetLimited(rateLimitErr.UntilReset())

		return
	}

	_ = t.errBudget.Take(1)
}
