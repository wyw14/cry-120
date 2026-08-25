package countdown

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Recycler struct {
	states  *StateStore
	journal *journal.Store
}

func NewRecycler(states *StateStore, store *journal.Store) *Recycler {
	return &Recycler{states: states, journal: store}
}

func (r *Recycler) Recycle(ctx context.Context, seconds int, now time.Time) (model.CountdownState, error) {
	current := r.states.Current()
	next := current
	next.Generation = model.NewIdentity("countdown")
	next.Phase = model.CountdownHeld
	next.TMinusSeconds = seconds
	next.StableUntil = time.Time{}
	next.PermitRevision = ""
	next.UpdatedAt = now.UTC()
	event, err := journal.NewEvent("countdown.recycled", next.Generation.String(), model.NewRevision(), map[string]string{"previous_generation": current.Generation.String()}, now)
	if err != nil {
		return model.CountdownState{}, err
	}
	if err := r.journal.Append(ctx, event); err != nil {
		return model.CountdownState{}, err
	}
	return r.states.Update(next), nil
}

func (r *Recycler) Generation() model.Identity {
	return r.states.Current().Generation
}
