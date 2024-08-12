package routing

import (
	"strconv"
	"testing"
	"time"

	"github.com/EinStack/glide/pkg/extmodel"

	"github.com/stretchr/testify/require"
)

func TestLeastLatencyRouting_Warmup(t *testing.T) {
	type Model struct {
		modelID string
		healthy bool
		latency float64
	}

	type TestCase struct {
		models           []Model
		expectedModelIDs []string
	}

	tests := map[string]TestCase{
		"all cold models":             {[]Model{{"first", true, 0.0}, {"second", true, 0.0}, {"third", true, 0.0}}, []string{"first", "second", "third"}},
		"all cold models & unhealthy": {[]Model{{"first", true, 0.0}, {"second", false, 0.0}, {"third", true, 0.0}}, []string{"first", "third", "first"}},
		"some models are warmed":      {[]Model{{"first", true, 100.0}, {"second", true, 0.0}, {"third", true, 120.0}}, []string{"second", "second", "second"}},
		"cold unhealthy model":        {[]Model{{"first", true, 120.0}, {"second", false, 0.0}, {"third", true, 100.0}}, []string{"third", "third", "third"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modelPool := make([]extmodel.Interface, 0, len(tc.models))

			for _, model := range tc.models {
				modelPool = append(modelPool, extmodel.NewLangModelMock(model.modelID, model.healthy, model.latency, 1))
			}

			routing := NewLeastLatencyRouting(extmodel.ChatMockLatency, modelPool)
			iterator := routing.Iterator()

			// loop three times over the whole pool to check if we return back to the begging of the list
			for _, modelID := range tc.expectedModelIDs {
				model, err := iterator.Next()

				require.NoError(t, err)
				require.Equal(t, modelID, model.ID())
			}
		})
	}
}

func TestLeastLatencyRouting_Routing(t *testing.T) {
	type Model struct {
		modelID  string
		healthy  bool
		latency  float64
		expireAt time.Time
	}

	type TestCase struct {
		models           []Model
		expectedModelIDs []string
	}

	tests := map[string]TestCase{
		"no cold expired models": {
			[]Model{
				{"first", true, 100.0, time.Now().Add(30 * time.Second)},
				{"second", true, 80.0, time.Now().Add(30 * time.Second)},
				{"third", true, 101.0, time.Now().Add(30 * time.Second)},
			},
			[]string{"second", "second", "second"},
		},
		"one expired model": {
			[]Model{
				{"first", true, 100.0, time.Now().Add(30 * time.Second)},
				{"second", true, 80.0, time.Now().Add(30 * time.Second)},
				{"third", true, 101.0, time.Now().Add(-30 * time.Second)},
			},
			[]string{"third", "second", "second"},
		},
		"two expired models": {
			[]Model{
				{"first", true, 100.0, time.Now().Add(-60 * time.Second)},
				{"second", true, 80.0, time.Now().Add(30 * time.Second)},
				{"third", true, 101.0, time.Now().Add(-30 * time.Second)},
			},
			[]string{"first", "third", "second"},
		},
		"all expired models": {
			[]Model{
				{"first", true, 100.0, time.Now().Add(-30 * time.Second)},
				{"second", true, 80.0, time.Now().Add(-20 * time.Second)},
				{"third", true, 101.0, time.Now().Add(-60 * time.Second)},
			},
			[]string{"third", "first", "second"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			schedules := make([]*ModelSchedule, 0, len(tc.models))

			for _, model := range tc.models {
				schedules = append(schedules, &ModelSchedule{
					model: extmodel.NewLangModelMock(
						model.modelID,
						model.healthy,
						model.latency,
						1,
					),
					expireAt: model.expireAt,
				})
			}

			routing := LeastLatencyRouting{
				latencyGetter: extmodel.ChatMockLatency,
				schedules:     schedules,
			}

			iterator := routing.Iterator()

			// loop three times over the whole pool to check if we return back to the begging of the list
			for _, modelID := range tc.expectedModelIDs {
				model, err := iterator.Next()

				require.NoError(t, err)
				require.Equal(t, modelID, model.ID())
			}
		})
	}
}

func TestLeastLatencyRouting_NoHealthyModels(t *testing.T) {
	tests := map[string][]float64{
		"all cold models unhealthy":    {0.0, 0.0, 0.0},
		"all warm models unhealthy":    {100.0, 120.0, 150.0},
		"cold & warm models unhealthy": {0.0, 120.0, 150.0},
	}

	for name, latencies := range tests {
		t.Run(name, func(t *testing.T) {
			modelPool := make([]extmodel.Interface, 0, len(latencies))

			for idx, latency := range latencies {
				modelPool = append(modelPool, extmodel.NewLangModelMock(strconv.Itoa(idx), false, latency, 1))
			}

			routing := NewLeastLatencyRouting(extmodel.ChatLatency, modelPool)
			iterator := routing.Iterator()

			_, err := iterator.Next()
			require.ErrorIs(t, err, ErrNoHealthyModels)
		})
	}
}
