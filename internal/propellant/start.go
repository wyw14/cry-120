package propellant

import (
	"context"
	"errors"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
)

type StartService struct {
	sessions *SessionStore
	routes   *manifold.RouteService
	events   *journal.Store
}

func NewStartService(sessions *SessionStore, routes *manifold.RouteService, events *journal.Store) *StartService {
	return &StartService{sessions: sessions, routes: routes, events: events}
}

func (s *StartService) Start(ctx context.Context, kind model.PropellantKind, arm string, now time.Time) (model.FillSession, error) {
	if kind != model.Fuel && kind != model.Oxidizer {
		return model.FillSession{}, errors.New("unsupported propellant")
	}
	if current, ok := s.sessions.Current(kind); ok && !current.Terminal() {
		return model.FillSession{}, model.ErrConflict
	}
	session := model.NewFillSession(kind, arm, now)
	path, err := s.routes.Open(ctx, arm, session.ID, now)
	if err != nil {
		return model.FillSession{}, err
	}
	session.RouteToken = path.Token
	session.Phase = model.FillChilldown
	event, err := journal.NewEvent("fill.started", session.ID.String(), model.NewRevision(), session, now)
	if err != nil {
		s.routes.Close(arm, session.ID)
		return model.FillSession{}, err
	}
	if err := s.events.Append(ctx, event); err != nil {
		s.routes.Close(arm, session.ID)
		return model.FillSession{}, err
	}
	s.sessions.Update(session)
	return session, nil
}

func (s *StartService) StartResult(ctx context.Context, operation model.OperationID, kind model.PropellantKind, arm string, now time.Time) (model.Result, model.FillSession) {
	session, err := s.Start(ctx, kind, arm, now)
	if errors.Is(err, model.ErrConflict) {
		return model.Rejected(operation, "manifold_busy", "shared transfer manifold is occupied", true), model.FillSession{}
	}
	if err != nil {
		return model.Rejected(operation, "start_failed", err.Error(), true), model.FillSession{}
	}
	return model.Accepted(operation, "fill session started"), session
}
