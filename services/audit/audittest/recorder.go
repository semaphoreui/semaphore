package audittest

import (
	"context"
	"fmt"
	"sync"

	"github.com/semaphoreui/semaphore/services/audit"
)

type Recorded struct {
	Event   audit.Event
	Actor   audit.Actor
	Request audit.RequestInfo
}

type Recorder struct {
	mu       sync.Mutex
	recorded []Recorded
}

func (r *Recorder) Record(ctx context.Context, event audit.Event) {
	if event.Outcome == "" {
		event.Outcome = audit.OutcomeSuccess
	}
	request, _ := audit.RequestFrom(ctx)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = append(r.recorded, Recorded{Event: event, Actor: audit.ActorFrom(ctx), Request: request})
}

func (r *Recorder) All() []Recorded {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Recorded(nil), r.recorded...)
}

// Only also fails on a catalog violation.
func (r *Recorder) Only(kind audit.Kind) (Recorded, error) {
	all := r.All()
	if len(all) != 1 {
		return Recorded{}, fmt.Errorf("want one %s event, recorded %v", kind, kindsOf(all))
	}
	if all[0].Event.Kind != kind {
		return Recorded{}, fmt.Errorf("want %s, recorded %s", kind, all[0].Event.Kind)
	}
	return all[0], audit.Validate(all[0].Event)
}

func (r *Recorder) Kinds() ([]audit.Kind, error) {
	all := r.All()
	for _, recorded := range all {
		if err := audit.Validate(recorded.Event); err != nil {
			return nil, err
		}
	}
	return kindsOf(all), nil
}

func kindsOf(all []Recorded) []audit.Kind {
	kinds := make([]audit.Kind, 0, len(all))
	for _, recorded := range all {
		kinds = append(kinds, recorded.Event.Kind)
	}
	return kinds
}
