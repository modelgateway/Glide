package cohere

import (
	"strings"

	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/api/schemas"
	"go.uber.org/zap"
)

var (
	// Reference: https://platform.openai.com/docs/api-reference/chat/object
	CompleteReason  = "complete"
	MaxTokensReason = "max_tokens"
	FilteredReason  = "error_toxic"
	// TODO: How to process  ERROR_LIMIT & ERROR?
)

func NewFinishReasonMapper(tel *telemetry.Telemetry) *FinishReasonMapper {
	return &FinishReasonMapper{
		tel: tel,
	}
}

type FinishReasonMapper struct {
	tel *telemetry.Telemetry
}

func (m *FinishReasonMapper) Map(finishReason *string) *schemas.FinishReason {
	if finishReason == nil || len(*finishReason) == 0 {
		return nil
	}

	var reason *schemas.FinishReason

	switch strings.ToLower(*finishReason) {
	case CompleteReason:
		reason = &schemas.ReasonComplete
	case MaxTokensReason:
		reason = &schemas.ReasonMaxTokens
	case FilteredReason:
		reason = &schemas.ReasonContentFiltered
	default:
		m.tel.Logger.Warn(
			"Unknown finish reason, other is going to used",
			zap.String("unknown_reason", *finishReason),
		)

		reason = &schemas.ReasonOther
	}

	return reason
}
