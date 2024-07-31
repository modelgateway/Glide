package provider

import (
	"github.com/EinStack/glide/pkg/clients"
	"github.com/EinStack/glide/pkg/telemetry"
)

// TODO: ProviderConfig should be more generic, not tied to LangProviders
type ProviderConfig interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error
	ToClient(tel *telemetry.Telemetry, clientConfig *clients.ClientConfig) (LangProvider, error)
}
