package interlock

import (
	"context"
	"sync"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Recovery struct {
	mu        sync.Mutex
	store     *journal.SnapshotStore
	holds     *HoldAggregate
	permits   *PermitStore
	recovered bool
}

type recoveredState struct {
	Holds   []model.Hold   `json:"holds"`
	Permits []model.Permit `json:"permits"`
}

func NewRecovery(store *journal.SnapshotStore, holds *HoldAggregate, permits *PermitStore) *Recovery {
	return &Recovery{store: store, holds: holds, permits: permits}
}

func (r *Recovery) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var state recoveredState
	_, err := r.store.Load("interlock", &state)
	if err != nil {
		return err
	}
	r.holds.Replace(state.Holds)
	r.permits.Restore(state.Permits)
	r.mu.Lock()
	r.recovered = true
	r.mu.Unlock()
	return nil
}

func (r *Recovery) Save() error {
	return r.store.Save("interlock", recoveredState{Holds: r.holds.Active(), Permits: r.permits.List()})
}

func (r *Recovery) Ready() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recovered
}
