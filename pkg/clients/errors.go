package clients

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrEmptyResponse            = errors.New("empty model response")
	ErrProviderUnavailable      = errors.New("provider is not available")
	ErrUnauthorized             = errors.New("API key is wrong or not set")
	ErrChatStreamNotImplemented = errors.New("streaming chat API is not implemented for provider")
)

type RateLimitError struct {
	untilReset time.Duration
}

func (e RateLimitError) Error() string {
	return fmt.Sprintf("rate limit reached, please wait %v", e.untilReset)
}

func (e RateLimitError) UntilReset() time.Duration {
	return e.untilReset
}

func NewRateLimitError(untilReset *time.Duration) *RateLimitError {
	defaultResetTime := 1 * time.Minute

	if untilReset == nil {
		untilReset = &defaultResetTime
	}

	return &RateLimitError{
		untilReset: *untilReset,
	}
}
