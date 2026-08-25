package propellant

import (
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
)

type Status struct {
	Sessions []model.FillSession  `json:"sessions"`
	Routes   []manifold.PathState `json:"routes"`
	Leases   []manifold.Lease     `json:"leases"`
}

type StatusService struct {
	sessions *SessionStore
	routes   *manifold.RouteService
	leases   *manifold.LeaseManager
}

func NewStatusService(sessions *SessionStore, routes *manifold.RouteService, leases *manifold.LeaseManager) *StatusService {
	return &StatusService{sessions: sessions, routes: routes, leases: leases}
}

func (s *StatusService) Current() Status {
	return Status{Sessions: s.sessions.List(), Routes: s.routes.List(), Leases: s.leases.All()}
}
