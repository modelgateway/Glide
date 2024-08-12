package clients

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/EinStack/glide/pkg/config/fields"
)

func TestClientConfig_DefaultConfig(t *testing.T) {
	config := DefaultClientConfig()

	require.NotEmpty(t, config.Timeout)
}

func TestClientConfig_JSONMarshal(t *testing.T) {
	defaultConfig := DefaultClientConfig()

	expectedJSON := `{
        "timeout": "10s",
		"max_idle_connections": 100,
		"max_idle_connections_per_host": 2	
	}`

	marshaledJSON, err := json.MarshalIndent(defaultConfig, "", "\t")

	require.NoError(t, err)
	require.JSONEq(t, expectedJSON, string(marshaledJSON))
}

func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig()

	require.NotNil(t, config, "Config must not be nil")
	require.NotNil(t, config.Timeout, "Timeout must not be nil")
	require.NotNil(t, config.MaxIdleConns, "MaxIdleConns must not be nil")
	require.NotNil(t, config.MaxIdleConnsPerHost, "MaxIdleConnsPerHost must not be nil")

	// Check default timeout
	expectedTimeout := fields.Duration(10 * time.Second)
	require.Equal(t, expectedTimeout, *config.Timeout)

	// Check MaxIdleConns
	expectedMaxIdleConns := 100
	require.Equal(t, expectedMaxIdleConns, *config.MaxIdleConns)

	// Check MaxIdleConnsPerHost
	expectedMaxIdleConnsPerHost := 2
	require.Equal(t, expectedMaxIdleConnsPerHost, *config.MaxIdleConnsPerHost)
}
