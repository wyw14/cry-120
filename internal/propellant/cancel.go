package propellant

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Safing interface {
	SafeSession(context.Context, model.FillSession, time.Time) error
}

type OperationGate interface {
	AllowOperations(bool)
}

type CancelService struct {
	sessions *SessionStore
	safing   Safing
	gate     OperationGate
	events   *journal.Store
}

func NewCancelService(sessions *SessionStore, safing Safing, gate OperationGate, events *journal.Store) *CancelService {
	return &CancelService{sessions: sessions, safing: safing, gate: gate, events: events}
}

func (s *CancelService) Cancel(ctx context.Context, sessionID model.Identity, now time.Time) (model.FillSession, error) {
	session, ok := s.sessions.Get(sessionID)
	if !ok {
		return model.FillSession{}, model.ErrNotFound
	}
	if session.Terminal() {
		return session, nil
	}
	session.Phase = model.FillCancelling
	session.UpdatedAt = now.UTC()
	s.sessions.Update(session)
	// Gate off new operations while safing is in flight. The safing worker
	// (cutoff, drain, route-release) must finish before the cancel is allowed
	// to complete externally; otherwise a stale worker from this session can
	// later close an isolation valve that a freshly-started session reopened.
	s.gate.AllowOperations(false)
	// Run safing synchronously on a detached context so a client disconnect
	// cannot abort safety-critical closure. SafeSession does not return until
	// every safing step has exited, which guarantees no stale worker outlives
	// this call.
	if err := s.safing.SafeSession(context.Background(), session, now); err != nil {
		// Safing failed: the worker already published the "pad not safe" hold
		// and disabled operations. Leave the session in cancelling and keep
		// operations gated so a new fill cannot start against an unsafe pad.
		return session, err
	}
	event, err := journal.NewEvent("fill.cancelled", session.ID.String(), model.NewRevision(), session, now)
	if err != nil {
		return session, err
	}
	if err := s.events.Append(ctx, event); err != nil {
		return session, err
	}
	session.Phase = model.FillCancelled
	session.UpdatedAt = time.Now().UTC()
	s.sessions.Update(session)
	// All safing actions have exited; it is now safe to let new operations in.
	s.gate.AllowOperations(true)
	return session, nil
}

func (s *CancelService) Status(sessionID model.Identity) (model.FillSession, bool) {
	return s.sessions.Get(sessionID)
}
