package clients

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRateLimitError(t *testing.T) {
	duration := 5 * time.Minute
	err := NewRateLimitError(&duration)

	require.Equal(t, duration, err.UntilReset())
	require.Contains(t, err.Error(), "rate limit reached")
}
