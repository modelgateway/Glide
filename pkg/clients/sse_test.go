package clients

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSSEvent_ValidEvents(t *testing.T) {
	tests := []struct {
		name   string
		rawMsg string
		data   string
	}{
		{"data only", "data: {\"id\":\"chatcmpl-8wFR3h2Spa9XeRbipfaJczj42pZQg\"}\n", "{\"id\":\"chatcmpl-8wFR3h2Spa9XeRbipfaJczj42pZQg\"}"},
		{"empty data", "data:", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := ParseSSEvent([]byte(tt.rawMsg))

			require.NoError(t, err)
			require.Equal(t, []byte(tt.data), event.Data)
		})
	}
}
