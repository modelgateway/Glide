package routing

import (
	"sync/atomic"

	"github.com/EinStack/glide/pkg/extmodel"
)

const (
	Priority Strategy = "priority"
)

// PriorityRouting routes request to the first healthy model defined in the routing config
//
//	Priority of models are defined as position of the model on the list
//	(e.g. the first model definition has the highest priority, then the second model definition and so on)
type PriorityRouting struct {
	models []extmodel.Interface
}

func NewPriority(models []extmodel.Interface) *PriorityRouting {
	return &PriorityRouting{
		models: models,
	}
}

func (r *PriorityRouting) Iterator() LangModelIterator {
	iterator := PriorityIterator{
		idx:    &atomic.Uint64{},
		models: r.models,
	}

	return iterator
}

type PriorityIterator struct {
	idx    *atomic.Uint64
	models []extmodel.Interface
}

func (r PriorityIterator) Next() (extmodel.Interface, error) {
	modelPool := r.models

	for idx := int(r.idx.Load()); idx < len(modelPool); idx = int(r.idx.Add(1)) {
		model := modelPool[idx]

		if !model.Healthy() {
			continue
		}

		return model, nil
	}

	return nil, ErrNoHealthyModels
}
