package cohere

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EinStack/glide/pkg/clients"

	"github.com/EinStack/glide/pkg/telemetry"

	"go.uber.org/zap"
)

type ErrorMapper struct {
	tel *telemetry.Telemetry
}

func NewErrorMapper(tel *telemetry.Telemetry) *ErrorMapper {
	return &ErrorMapper{
		tel: tel,
	}
}

func (m *ErrorMapper) Map(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		m.tel.Logger.Error(
			"Failed to unmarshal chat response error",
			zap.String("provider", ProviderID),
			zap.Error(err),
			zap.ByteString("rawResponse", bodyBytes),
		)

		return clients.ErrProviderUnavailable
	}

	m.tel.Logger.Error(
		"Chat request failed",
		zap.String("provider", ProviderID),
		zap.Int("statusCode", resp.StatusCode),
		zap.String("response", string(bodyBytes)),
		zap.Any("headers", resp.Header),
	)

	if resp.StatusCode == http.StatusTooManyRequests {
		// Read the value of the "Retry-After" header to get the cooldown delay
		retryAfter := resp.Header.Get("Retry-After")

		// Parse the value to get the duration
		cooldownDelay, err := time.ParseDuration(retryAfter)
		if err != nil {
			return fmt.Errorf("failed to parse cooldown delay from headers: %w", err)
		}

		return clients.NewRateLimitError(&cooldownDelay)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return clients.ErrUnauthorized
	}

	// Server & client errors result in the same error to keep gateway resilient
	return clients.ErrProviderUnavailable
}
