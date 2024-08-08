package ollama

import (
	"net/http"
	"net/url"
	"time"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"
)

const (
	providerName = "ollama"
)

// Client is a client for accessing OpenAI API
type Client struct {
	baseURL             string
	chatURL             string
	chatRequestTemplate *ChatRequest
	config              *Config
	httpClient          *http.Client
	telemetry           *telemetry.Telemetry
}

// NewClient creates a new OpenAI client for the OpenAI API.
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
