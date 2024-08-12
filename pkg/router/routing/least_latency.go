package routing

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/EinStack/glide/pkg/extmodel"

	"github.com/EinStack/glide/pkg/router/latency"
)

const (
	LeastLatency Strategy = "least_latency"
)

// LatencyGetter defines where to find latency for the specific model action
type LatencyGetter = func(model extmodel.Interface) *latency.MovingAverage

// ModelSchedule defines latency update schedule for models
type ModelSchedule struct {
	mu       sync.RWMutex
	model    extmodel.Interface
	expireAt time.Time
}

func NewSchedule(model extmodel.Interface) *ModelSchedule {
	schedule := &ModelSchedule{
		model: model,
	}

	schedule.Update()

	return schedule
}

func (s *ModelSchedule) ExpireAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.expireAt
}

func (s *ModelSchedule) Expired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Now().After(s.expireAt)
}

// Update expands the expiration deadline
func (s *ModelSchedule) Update() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.expireAt = time.Now().Add(time.Duration(*s.model.LatencyUpdateInterval()))
}

// LeastLatencyRouting routes requests to the model that responses the fastest
// At the beginning, we try to send requests to all models to find out the quickest one.
// After that, we use the that model for some time. But we don't want to stick to that model forever (as some
// other model latency may improve over time overperform the best one),
// so we need to send some traffic to other models from time to time to update their latency stats
type LeastLatencyRouting struct {
	latencyGetter LatencyGetter
	warmupIdx     atomic.Uint32
	schedules     []*ModelSchedule
}

func NewLeastLatencyRouting(latencyGetter LatencyGetter, models []extmodel.Interface) *LeastLatencyRouting {
	schedules := make([]*ModelSchedule, 0, len(models))

	for _, model := range models {
		schedules = append(schedules, NewSchedule(model))
	}

	return &LeastLatencyRouting{
		latencyGetter: latencyGetter,
		schedules:     schedules,
	}
}

func (r *LeastLatencyRouting) Iterator() LangModelIterator {
	return r
}

// Next picks a model with the least average latency over time
// The algorithm consists of two stages:
//   - warm up: Before considering model latencies we may want to collect more than one sample to make better decisions.
//     To learn about latencies, we route requests to all "cold" models in round-robin manner
//   - least latency selection: Once all models are warmed, we pick one with the least latency
//
// Additionally, we should update our stats as response latency is a dynamic distribution,
// we cannot simply stick to the fastest model discovered on the warmup stage (as we could overlook
// other model latencies that might have improved over time).
// For that, we introduced expiration time after which the model receives a request
// even if it was not the fastest to respond
func (r *LeastLatencyRouting) Next() (extmodel.Interface, error) { //nolint:cyclop
	coldSchedules := r.getColdModelSchedules()

	if len(coldSchedules) > 0 {
		// warm up models
		idx := r.warmupIdx.Add(1) - 1

		schedule := coldSchedules[idx%uint32(len(coldSchedules))]
		schedule.Update()

		return schedule.model, nil
	}

	// latency-based routing
	var nextSchedule *ModelSchedule

	for _, schedule := range r.schedules {
		if !schedule.model.Healthy() {
			// cannot do much with unavailable model
			continue
		}

		if nextSchedule == nil {
			nextSchedule = schedule
			continue
		}

		// We pick either the earliest expired model or one with the least response latency

		if schedule.Expired() && schedule.ExpireAt().Before(nextSchedule.ExpireAt()) {
			// if the model latency is expired, then it should be picked only if
			//  it's expiration time happened earlier than the prev picked model
			nextSchedule = schedule
			continue
		}

		if !schedule.Expired() && !nextSchedule.Expired() &&
			r.latencyGetter(nextSchedule.model).Value() > r.latencyGetter(schedule.model).Value() {
			nextSchedule = schedule
		}
	}

	if nextSchedule != nil {
		nextSchedule.Update()

		return nextSchedule.model, nil
	}

	return nil, ErrNoHealthyModels
}

func (r *LeastLatencyRouting) getColdModelSchedules() []*ModelSchedule {
	coldModels := make([]*ModelSchedule, 0, len(r.schedules))

	for _, schedule := range r.schedules {
		if schedule.model.Healthy() && !r.latencyGetter(schedule.model).WarmedUp() {
			coldModels = append(coldModels, schedule)
		}
	}

	return coldModels
}
