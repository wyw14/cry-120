package countdown

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type StateStore struct {
	mu      sync.RWMutex
	state   model.CountdownState
	history []model.CountdownState
}

func NewStateStore(seconds int, now time.Time) *StateStore {
	initial := model.NewCountdown(seconds, now)
	return &StateStore{state: initial, history: []model.CountdownState{initial}}
}

func (s *StateStore) Current() model.CountdownState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Clone()
}

func (s *StateStore) Update(next model.CountdownState) model.CountdownState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = next.Clone()
	s.history = append(s.history, next.Clone())
	return s.state.Clone()
}

func (s *StateStore) Replace(next model.CountdownState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = next.Clone()
	s.history = []model.CountdownState{next.Clone()}
}

func (s *StateStore) History() []model.CountdownState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.CountdownState, len(s.history))
	copy(result, s.history)
	return result
}

func (s *StateStore) ActiveGenerations() []model.Identity {
	items := s.History()
	seen := make(map[model.Identity]struct{})
	result := []model.Identity{}
	for _, state := range items {
		if state.Phase == model.CountdownRunning {
			if _, ok := seen[state.Generation]; !ok {
				seen[state.Generation] = struct{}{}
				result = append(result, state.Generation)
			}
		}
	}
	return result
}
