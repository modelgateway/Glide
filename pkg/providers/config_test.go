package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDynLangProvider(t *testing.T) {
	LangRegistry.Register(ProviderTest, &TestConfig{})

	type ProviderConfig struct {
		Provider *Config `yaml:"provider"`
	}

	prConfig := make(Config)
	providerConfig := ProviderConfig{
		Provider: &prConfig,
	}

	config, err := os.ReadFile(filepath.Clean("./config_test.yaml"))
	require.NoError(t, err)

	err = yaml.Unmarshal(config, &providerConfig)
	require.NoError(t, err)
}
