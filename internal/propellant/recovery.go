package propellant

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
)

type Recovery struct {
	mu        sync.Mutex
	snapshots *journal.SnapshotStore
	sessions  *SessionStore
	flow      *manifold.Recovery
	commands  *manifold.Commander
	recovered bool
}

func NewRecovery(snapshots *journal.SnapshotStore, sessions *SessionStore, flow *manifold.Recovery, commands *manifold.Commander) *Recovery {
	return &Recovery{snapshots: snapshots, sessions: sessions, flow: flow, commands: commands}
}

func (r *Recovery) RecoveryName() string {
	return "propellant"
}

func (r *Recovery) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !r.flow.Ready() {
		return model.ErrOperationPending
	}
	var sessions []model.FillSession
	_, err := r.snapshots.Load("propellant", &sessions)
	if err != nil {
		return err
	}
	r.sessions.Replace(sessions)
	for _, session := range sessions {
		if session.Terminal() {
			continue
		}
		if _, err := r.commands.Issue(ctx, "transfer-manifold", session.Arm+"-flow", "replenish", session.ID, session.RouteToken, time.Now()); err != nil {
			return err
		}
	}
	r.mu.Lock()
	r.recovered = true
	r.mu.Unlock()
	return nil
}

func (r *Recovery) Save() error {
	return r.snapshots.Save("propellant", r.sessions.List())
}

func (r *Recovery) Ready() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recovered
}
