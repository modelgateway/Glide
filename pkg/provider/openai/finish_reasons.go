package openai

import (
	"github.com/EinStack/glide/pkg/api/schema"
	"github.com/EinStack/glide/pkg/telemetry"

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

func (m *FinishReasonMapper) Map(finishReason string) *schema.FinishReason {
	if len(finishReason) == 0 {
		return nil
	}

	var reason *schema.FinishReason

	switch finishReason {
	case CompleteReason:
		reason = &schema.ReasonComplete
	case MaxTokensReason:
		reason = &schema.ReasonMaxTokens
	case FilteredReason:
		reason = &schema.ReasonContentFiltered
	default:
		m.tel.Logger.Warn(
			"Unknown finish reason, other is going to used",
			zap.String("unknown_reason", finishReason),
		)

		reason = &schema.ReasonOther
	}

	return reason
}
