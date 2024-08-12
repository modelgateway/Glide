package health

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const budgetSeparator = "/"

type Unit string

const (
	MILLI Unit = "ms"
	MIN   Unit = "m"
	SEC   Unit = "s"
	HOUR  Unit = "h"
)

// ErrorBudget parses human-friendly error budget representation and return it as errors & update rate pair
// Error budgets could be set as a string in the following format: "10/s", "5/ms", "100/m" "1500/h"
type ErrorBudget struct {
	budget uint
	unit   Unit
}

func NewErrorBudget(budget uint, unit Unit) *ErrorBudget {
	return &ErrorBudget{
		budget: budget,
		unit:   unit,
	}
}

func DefaultErrorBudget() *ErrorBudget {
	return &ErrorBudget{
		budget: 10,
		unit:   MIN,
	}
}

// Budget defines max allows number of errors per given time period
func (b *ErrorBudget) Budget() uint {
	return b.budget
}

// TimePerTokenMicro defines how much time do we need to wait to get one error token recovered (in microseconds)
func (b *ErrorBudget) TimePerTokenMicro() uint {
	return b.unitToMicro(b.unit) / b.budget
}

// MarshalText implements the encoding.TextMarshaler interface.
// This marshals the type and name as one string in the config.
func (b *ErrorBudget) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

func (b *ErrorBudget) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), budgetSeparator)

	if len(parts) != 2 {
		return errors.New("invalid format")
	}

	budget, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("error parsing error number: %v", err)
	}

	if budget <= 0 {
		return fmt.Errorf("error number should be greater then 0 (%v given)", budget)
	}

	unit := Unit(parts[1])

	if unit != MILLI && unit != SEC && unit != MIN && unit != HOUR {
		return errors.New("invalid unit (supported: ms, s, m, h)")
	}

	b.budget = uint(budget)
	b.unit = unit

	return nil
}

func (b *ErrorBudget) unitToMicro(unit Unit) uint {
	switch unit {
	case MILLI:
		return 1_000 // 1 ms = 1000 microseconds
	case SEC:
		return 1_000_000 // 1 s = 1,000,000 microseconds
	case MIN:
		return 60_000_000 // 1 m = 60,000,000 microseconds
	case HOUR:
		return 3_600_000_000 // 1 h = 3,600,000,000 microseconds
	default:
		return 1
	}
}

// String returns the ID string representation as "type[/name]" format.
func (b *ErrorBudget) String() string {
	return strconv.Itoa(int(b.budget)) + budgetSeparator + string(b.unit)
}
