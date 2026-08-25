package manifold

import (
	"context"
	"sync"

	"github.com/wyw14/cry-120/internal/journal"
)

type Recovery struct {
	mu        sync.Mutex
	store     *journal.SnapshotStore
	leases    *LeaseManager
	recovered bool
}

func NewRecovery(store *journal.SnapshotStore, leases *LeaseManager) *Recovery {
	return &Recovery{store: store, leases: leases}
}

func (r *Recovery) RecoveryName() string {
	return "manifold"
}

func (r *Recovery) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var leases []Lease
	_, err := r.store.Load("manifold-leases", &leases)
	if err != nil {
		return err
	}
	r.leases.Restore(leases)
	r.mu.Lock()
	r.recovered = true
	r.mu.Unlock()
	return nil
}

func (r *Recovery) Save() error {
	return r.store.Save("manifold-leases", r.leases.All())
}

func (r *Recovery) Ready() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recovered
}
