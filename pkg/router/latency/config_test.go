package latency

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatencyConfig_Default(t *testing.T) {
	config := DefaultConfig()

	require.NotEmpty(t, config)
}
