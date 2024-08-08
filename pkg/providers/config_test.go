package providers

import (
	"os"
	"path/filepath"
	"testing"

	testprovider "github.com/EinStack/glide/pkg/providers/testing"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDynLangProvider(t *testing.T) {
	LangRegistry.Register(testprovider.ProviderTest, &testprovider.Config{})

	type ProviderConfig struct {
		Provider *DynLangProvider `yaml:"provider"`
	}

	prConfig := make(DynLangProvider)
	providerConfig := ProviderConfig{
		Provider: &prConfig,
	}

	config, err := os.ReadFile(filepath.Clean("./config_test.yaml"))
	require.NoError(t, err)

	err = yaml.Unmarshal(config, &providerConfig)
	require.NoError(t, err)
}
