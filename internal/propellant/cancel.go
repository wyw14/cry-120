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
	s.gate.AllowOperations(false)
	if err := s.safing.SafeSession(ctx, session, now); err != nil {
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
	s.gate.AllowOperations(true)
	return session, nil
}

func (s *CancelService) Status(sessionID model.Identity) (model.FillSession, bool) {
	return s.sessions.Get(sessionID)
}
