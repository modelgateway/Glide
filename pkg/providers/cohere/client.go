package cohere

import (
	"net/http"
	"net/url"
	"time"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"
)

const (
	providerName = "cohere"
)

// Client is a client for accessing Cohere API
type Client struct {
	baseURL             string
	chatURL             string
	chatRequestTemplate *ChatRequest
	finishReasonMapper  *FinishReasonMapper
	errMapper           *ErrorMapper
	config              *Config
	httpClient          *http.Client
	tel                 *telemetry.Telemetry
}

// NewClient creates a new Cohere client for the Cohere API.
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
		errMapper:          NewErrorMapper(tel),
		finishReasonMapper: NewFinishReasonMapper(tel),
		tel:                tel,
	}

	return c, nil
}

func (c *Client) Provider() string {
	return providerName
}

func (c *Client) ModelName() string {
	return c.config.ModelName
}
