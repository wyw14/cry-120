package manifold

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type PathState struct {
	Name       string         `json:"name"`
	Owner      model.Identity `json:"owner"`
	Token      model.Token    `json:"token"`
	LiquidOpen bool           `json:"liquid_open"`
	VaporOpen  bool           `json:"vapor_open"`
	Balanced   bool           `json:"balanced"`
}

type RouteService struct {
	mu     sync.Mutex
	paths  map[string]PathState
	leases *LeaseManager
}

func NewRouteService(leases *LeaseManager) *RouteService {
	return &RouteService{paths: make(map[string]PathState), leases: leases}
}

func (s *RouteService) Open(ctx context.Context, arm string, owner model.Identity, now time.Time) (PathState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, err := s.leases.Acquire(ctx, "transfer-manifold", owner, 15*time.Minute, now)
	if err != nil {
		return PathState{}, err
	}
	path := PathState{Name: arm, Owner: owner, Token: lease.Token, LiquidOpen: true, VaporOpen: true, Balanced: true}
	s.paths[arm] = path
	return path, nil
}

func (s *RouteService) Switch(ctx context.Context, oldArm, newArm string, owner model.Identity, now time.Time) (PathState, error) {
	if err := ctx.Err(); err != nil {
		return PathState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.paths[oldArm]
	if !ok || old.Owner != owner {
		return PathState{}, model.ErrNotFound
	}
	old.LiquidOpen = false
	old.VaporOpen = false
	old.Balanced = true
	s.paths[oldArm] = old
	current, ok := s.leases.Current("transfer-manifold")
	if !ok || current.Owner != owner || current.Token != old.Token {
		return PathState{}, fmt.Errorf("route lease lost: %w", model.ErrConflict)
	}
	next := PathState{Name: newArm, Owner: owner, Token: current.Token, LiquidOpen: true, VaporOpen: true, Balanced: true}
	s.paths[newArm] = next
	return next, nil
}

func (s *RouteService) Close(arm string, owner model.Identity) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, ok := s.paths[arm]
	if !ok || path.Owner != owner {
		return false
	}
	path.LiquidOpen = false
	path.VaporOpen = false
	path.Balanced = true
	s.paths[arm] = path
	return true
}

func (s *RouteService) Get(arm string) (PathState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, ok := s.paths[arm]
	return path, ok
}

func (s *RouteService) List() []PathState {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]PathState, 0, len(s.paths))
	for _, path := range s.paths {
		result = append(result, path)
	}
	return result
}
