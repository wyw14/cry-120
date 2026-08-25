package propellant

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
)

type SwitchService struct {
	sessions *SessionStore
	routes   *manifold.RouteService
	events   *journal.Store
}

func NewSwitchService(sessions *SessionStore, routes *manifold.RouteService, events *journal.Store) *SwitchService {
	return &SwitchService{sessions: sessions, routes: routes, events: events}
}

func (s *SwitchService) Switch(ctx context.Context, sessionID model.Identity, newArm string, now time.Time) (model.FillSession, error) {
	session, ok := s.sessions.Get(sessionID)
	if !ok {
		return model.FillSession{}, model.ErrNotFound
	}
	if session.Phase != model.FillChilldown {
		return model.FillSession{}, model.ErrInvalidTransition
	}
	oldArm := session.Arm
	path, err := s.routes.Switch(ctx, oldArm, newArm, session.ID, now)
	if err != nil {
		return model.FillSession{}, err
	}
	oldPath, ok := s.routes.Get(oldArm)
	if !ok || oldPath.LiquidOpen || oldPath.VaporOpen || !oldPath.Balanced {
		return model.FillSession{}, fmt.Errorf("old route not isolated: %w", model.ErrConflict)
	}
	if !path.LiquidOpen || !path.VaporOpen || !path.Balanced {
		return model.FillSession{}, fmt.Errorf("new route incomplete: %w", model.ErrConflict)
	}
	session.Arm = newArm
	session.RouteToken = path.Token
	session.UpdatedAt = now.UTC()
	event, err := journal.NewEvent("fill.arm.switched", session.ID.String(), model.NewRevision(), map[string]string{"old_arm": oldArm, "new_arm": newArm}, now)
	if err != nil {
		return model.FillSession{}, err
	}
	if err := s.events.Append(ctx, event); err != nil {
		return model.FillSession{}, err
	}
	return s.sessions.Update(session), nil
}

func (s *SwitchService) Route(arm string) (manifold.PathState, bool) {
	return s.routes.Get(arm)
}
