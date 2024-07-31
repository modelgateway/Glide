package routing

import (
	ptesting "github.com/EinStack/glide/pkg/providers/testing"
	"testing"

	"github.com/EinStack/glide/pkg/models"

	"github.com/stretchr/testify/require"
)

func TestWRoundRobinRouting_RoutingDistribution(t *testing.T) {
	type Model struct {
		modelID string
		healthy bool
		weight  int
	}

	type TestCase struct {
		models       []Model
		numTries     int
		distribution map[string]int
	}

	tests := map[string]TestCase{
		"equal weights 1": {
			[]Model{
				{"first", true, 1},
				{"second", true, 1},
				{"three", true, 1},
			},
			999,
			map[string]int{
				"first":  333,
				"second": 333,
				"three":  333,
			},
		},
		"equal weights 2": {
			[]Model{
				{"first", true, 2},
				{"second", true, 2},
				{"three", true, 2},
			},
			999,
			map[string]int{
				"first":  333,
				"second": 333,
				"three":  333,
			},
		},
		"4-2 split": {
			[]Model{
				{"first", true, 4},
				{"second", true, 2},
				{"three", true, 2},
			},
			1000,
			map[string]int{
				"first":  500,
				"second": 250,
				"three":  250,
			},
		},
		"5-2-3 split": {
			[]Model{
				{"first", true, 2},
				{"second", true, 5},
				{"three", true, 3},
			},
			1000,
			map[string]int{
				"first":  200,
				"second": 500,
				"three":  300,
			},
		},
		"1-2-3 split": {
			[]Model{
				{"first", true, 1},
				{"second", true, 2},
				{"three", true, 3},
			},
			1000,
			map[string]int{
				"first":  167,
				"second": 333,
				"three":  500,
			},
		},
		"pareto split": {
			[]Model{
				{"first", true, 80},
				{"second", true, 20},
			},
			1000,
			map[string]int{
				"first":  800,
				"second": 200,
			},
		},
		"zero weight": {
			[]Model{
				{"first", true, 2},
				{"second", true, 0},
				{"three", true, 2},
			},
			1000,
			map[string]int{
				"first": 500,
				"three": 500,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modelPool := make([]models.Model, 0, len(tc.models))

			for _, model := range tc.models {
				modelPool = append(modelPool, ptesting.NewLangModelMock(model.modelID, model.healthy, 0, model.weight))
			}

			routing := NewWeightedRoundRobin(modelPool)
			iterator := routing.Iterator()

			actualDistribution := make(map[string]int, len(tc.models))

			// loop three times over the whole pool to check if we return back to the begging of the list
			for i := 0; i < tc.numTries; i++ {
				model, err := iterator.Next()

				require.NoError(t, err)

				actualDistribution[model.ID()]++
			}

			require.Equal(t, tc.distribution, actualDistribution)
		})
	}
}

func TestWRoundRobinRouting_NoHealthyModels(t *testing.T) {
	modelPool := []models.Model{
		ptesting.NewLangModelMock("first", false, 0, 1),
		ptesting.NewLangModelMock("second", false, 0, 2),
		ptesting.NewLangModelMock("third", false, 0, 3),
	}

	routing := NewWeightedRoundRobin(modelPool)
	iterator := routing.Iterator()

	_, err := iterator.Next()
	require.Error(t, err)
}
