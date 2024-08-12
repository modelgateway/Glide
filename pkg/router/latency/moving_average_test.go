package latency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMovingAverage_WarmUpAndAverage(t *testing.T) {
	latencies := []float64{100, 100, 150}
	movingAverage := NewMovingAverage(0.9, 3)

	for _, latency := range latencies {
		movingAverage.Add(latency)

		require.False(t, movingAverage.WarmedUp())
		require.InDelta(t, 0.0, movingAverage.Value(), 0.0001)
	}

	movingAverage.Add(160)

	require.True(t, movingAverage.WarmedUp())
	require.InDelta(t, 155.6667, movingAverage.Value(), 0.0001)

	movingAverage.Add(160)
	require.True(t, movingAverage.WarmedUp())
	require.InDelta(t, 159.5667, movingAverage.Value(), 0.0001)
}

func TestMovingAverage_SetValue(t *testing.T) {
	movingAverage := NewMovingAverage(0.9, 3)

	movingAverage.Set(200.0)

	require.True(t, movingAverage.WarmedUp())
	require.InDelta(t, 200.0, movingAverage.Value(), 0.0001)
}
