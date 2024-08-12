package health

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorBudget_ParseValidString(t *testing.T) {
	tests := map[string]struct {
		input        string
		errors       int
		unit         Unit
		timePerToken int
	}{
		"1/s":      {input: "1/s", errors: 1, unit: SEC},
		"10/ms":    {input: "10/ms", errors: 10, unit: MILLI},
		"1000/m":   {input: "1000/m", errors: 1000, unit: MIN},
		"100000/h": {input: "100000/h", errors: 100000, unit: HOUR},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			budget := DefaultErrorBudget()

			require.NoError(t, budget.UnmarshalText([]byte(tc.input)))
			require.Equal(t, tc.errors, int(budget.Budget()))
			require.Equal(t, tc.unit, budget.unit)
			require.Equal(t, tc.input, budget.String())
		})
	}
}

func TestErrorBudget_ParseInvalidString(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"0/s":    {input: "0/s"},
		"-1/s":   {input: "-1/s"},
		"1.9/s":  {input: "1.9/s"},
		"1,9/s":  {input: "1,9/s"},
		"100/d":  {input: "100/d"},
		"100/mo": {input: "100/mo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			budget := DefaultErrorBudget()

			require.Error(t, budget.UnmarshalText([]byte(tc.input)))
		})
	}
}
