package routing

import (
	"sync/atomic"

	"glide/pkg/providers"
)

const (
	Priority Strategy = "priority"
)

// PriorityRouting routes request to the first healthy model defined in the routing config
//
//	Priority of models are defined as position of the model on the list
//	(e.g. the first model definition has the highest priority, then the second model definition and so on)
type PriorityRouting struct {
	models []providers.Model
}

func NewPriorityRouting(models []providers.Model) *PriorityRouting {
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
	models []providers.Model
}

func (r PriorityIterator) Next() (providers.Model, error) {
	models := r.models
	idx := r.idx.Load()

	for int(idx) < len(models) {
		idx = r.idx.Load()
		model := models[idx]

		r.idx.Add(1)

		if model.Healthy() {
			return model, nil
		}
	}

	return nil, ErrNoHealthyModels
}
