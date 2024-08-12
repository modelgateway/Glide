package routing

import (
	"testing"

	"github.com/EinStack/glide/pkg/extmodel"

	"github.com/stretchr/testify/require"
)

func TestPriorityRouting_PickModelsInOrder(t *testing.T) {
	type Model struct {
		modelID string
		healthy bool
	}

	type TestCase struct {
		models           []Model
		expectedModelIDs []string
	}

	tests := map[string]TestCase{
		"all healthy":         {[]Model{{"first", true}, {"second", true}, {"third", true}}, []string{"first", "first", "first"}},
		"first unhealthy":     {[]Model{{"first", false}, {"second", true}, {"third", true}}, []string{"second", "second", "second"}},
		"first two unhealthy": {[]Model{{"first", false}, {"second", false}, {"third", true}}, []string{"third", "third", "third"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modelPool := make([]extmodel.Interface, 0, len(tc.models))

			for _, model := range tc.models {
				modelPool = append(modelPool, extmodel.NewLangModelMock(model.modelID, model.healthy, 100, 1))
			}

			routing := NewPriority(modelPool)
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

func TestPriorityRouting_NoHealthyModels(t *testing.T) {
	modelPool := []extmodel.Interface{
		extmodel.NewLangModelMock("first", false, 0, 1),
		extmodel.NewLangModelMock("second", false, 0, 1),
		extmodel.NewLangModelMock("third", false, 0, 1),
	}

	routing := NewPriority(modelPool)
	iterator := routing.Iterator()

	_, err := iterator.Next()
	require.Error(t, err)
}