package openai

import (
	"github.com/EinStack/glide/pkg/telemetry"

	"github.com/EinStack/glide/pkg/api/schemas"
	"go.uber.org/zap"
)

var (
	// Reference: https://platform.openai.com/docs/api-reference/chat/object

	CompleteReason  = "stop"
	MaxTokensReason = "length"
	FilteredReason  = "content_filter"
)

func NewFinishReasonMapper(tel *telemetry.Telemetry) *FinishReasonMapper {
	return &FinishReasonMapper{
		tel: tel,
	}
}

type FinishReasonMapper struct {
	tel *telemetry.Telemetry
}

func (m *FinishReasonMapper) Map(finishReason string) *schemas.FinishReason {
	if len(finishReason) == 0 {
		return nil
	}

	var reason *schemas.FinishReason

	switch finishReason {
	case CompleteReason:
		reason = &schemas.ReasonComplete
	case MaxTokensReason:
		reason = &schemas.ReasonMaxTokens
	case FilteredReason:
		reason = &schemas.ReasonContentFiltered
	default:
		m.tel.Logger.Warn(
			"Unknown finish reason, other is going to used",
			zap.String("unknown_reason", finishReason),
		)

		reason = &schemas.ReasonOther
	}

	return reason
}
