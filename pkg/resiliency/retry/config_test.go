package retry

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRetryConfig_DefaultConfig(t *testing.T) {
	config := DefaultExpRetryConfig()

	require.NotNil(t, config)
}

func TestRetryConfig_JSONMarshal(t *testing.T) {
	defaultConfig := DefaultExpRetryConfig()

	expectedJSON := `{
		"max_retries": 3,
		"base_multiplier": 2,
		"min_delay": "2s",
		"max_delay": "5s"
	}`

	marshaledJSON, err := json.MarshalIndent(defaultConfig, "", "\t")
	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(marshaledJSON))
}
