package health

import (
	"testing"
	"time"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/stretchr/testify/require"
)

func TestHealthTracker_HealthyByDefault(t *testing.T) {
	budget := NewErrorBudget(3, SEC)
	tracker := NewTracker(budget)

	require.True(t, tracker.Healthy())
}

func TestHealthTracker_UnhealthyWhenBugetExceeds(t *testing.T) {
	budget := NewErrorBudget(3, SEC)
	tracker := NewTracker(budget)

	for range 3 {
		tracker.TrackErr(clients.ErrProviderUnavailable)
	}

	require.False(t, tracker.Healthy())
}

func TestHealthTracker_RateLimited(t *testing.T) {
	budget := NewErrorBudget(3, SEC)
	tracker := NewTracker(budget)

	limitedUntil := 10 * time.Minute
	tracker.TrackErr(clients.NewRateLimitError(&limitedUntil))

	require.False(t, tracker.Healthy())
}
