package weather

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/model"
)

type HoldService struct {
	mu     sync.Mutex
	holds  *interlock.HoldAggregate
	active map[string]model.Hold
}

func NewHoldService(holds *interlock.HoldAggregate) *HoldService {
	return &HoldService{holds: holds, active: make(map[string]model.Hold)}
}

func (s *HoldService) Raise(reason string, now time.Time) model.Hold {
	s.mu.Lock()
	defer s.mu.Unlock()
	hold := s.holds.Publish("weather", reason, now)
	s.active[reason] = hold
	return hold
}

func (s *HoldService) Clear(reason string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.active[reason]; !ok {
		return false
	}
	if !s.holds.Release("", reason) {
		return false
	}
	delete(s.active, reason)
	return true
}

func (s *HoldService) Active() []model.Hold {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]model.Hold, 0, len(s.active))
	for _, hold := range s.active {
		result = append(result, hold)
	}
	return result
}

func (s *HoldService) Has(reason string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.active[reason]
	return ok
}
