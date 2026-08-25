package countdown

import (
	"context"
	"sync"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Recovery struct {
	mu        sync.Mutex
	snapshots *journal.SnapshotStore
	states    *StateStore
	ready     bool
}

func NewRecovery(snapshots *journal.SnapshotStore, states *StateStore) *Recovery {
	return &Recovery{snapshots: snapshots, states: states}
}

func (r *Recovery) RecoveryName() string {
	return "countdown"
}

func (r *Recovery) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var state model.CountdownState
	found, err := r.snapshots.Load("countdown", &state)
	if err != nil {
		return err
	}
	if found {
		r.states.Replace(state)
	}
	r.mu.Lock()
	r.ready = true
	r.mu.Unlock()
	return nil
}

func (r *Recovery) Save() error {
	return r.snapshots.Save("countdown", r.states.Current())
}

func (r *Recovery) Ready() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ready
}
