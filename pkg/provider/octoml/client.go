package octoml

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/EinStack/glide/pkg/provider"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"
)

const (
	providerName = "octoml"
)

// ErrEmptyResponse is returned when the OctoML API returns an empty response.
var (
	ErrEmptyResponse = errors.New("empty response")
)

// Client is a client for accessing OctoML API
type Client struct {
	baseURL             string
	chatURL             string
	chatRequestTemplate *ChatRequest
	errMapper           *ErrorMapper
	config              *Config
	httpClient          *http.Client
	telemetry           *telemetry.Telemetry
}

// ensure interfaces
var (
	_ provider.LangProvider = (*Client)(nil)
)

// NewClient creates a new OctoML client for the OctoML API.
func NewClient(providerConfig *Config, clientConfig *clients.ClientConfig, tel *telemetry.Telemetry) (*Client, error) {
	chatURL, err := url.JoinPath(providerConfig.BaseURL, providerConfig.ChatEndpoint)
	if err != nil {
		return nil, err
	}

	c := &Client{
		baseURL:             providerConfig.BaseURL,
		chatURL:             chatURL,
		config:              providerConfig,
		chatRequestTemplate: NewChatRequestFromConfig(providerConfig),
		errMapper:           NewErrorMapper(tel),
		httpClient: &http.Client{
			Timeout: time.Duration(*clientConfig.Timeout),
			Transport: &http.Transport{
				MaxIdleConns:        *clientConfig.MaxIdleConns,
				MaxIdleConnsPerHost: *clientConfig.MaxIdleConnsPerHost,
			},
		},
		telemetry: tel,
	}

	return c, nil
}

func (c *Client) Provider() string {
	return providerName
}

func (c *Client) ModelName() string {
	return c.config.ModelName
}