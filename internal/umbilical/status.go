package umbilical

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type ActionStore struct {
	mu      sync.RWMutex
	current model.UmbilicalAction
	history []model.UmbilicalAction
}

func NewActionStore(generation model.Identity, now time.Time) *ActionStore {
	action := model.UmbilicalAction{Token: model.NewToken("umbilical"), Generation: generation, State: model.UmbilicalConnected, UpdatedAt: now.UTC()}
	return &ActionStore{current: action, history: []model.UmbilicalAction{action}}
}

func (s *ActionStore) Current() model.UmbilicalAction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	current := s.current
	current.Controllers = append([]string{}, current.Controllers...)
	return current
}

func (s *ActionStore) Update(action model.UmbilicalAction) model.UmbilicalAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	action.Controllers = append([]string{}, action.Controllers...)
	s.current = action
	s.history = append(s.history, action)
	return action
}

func (s *ActionStore) History() []model.UmbilicalAction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.UmbilicalAction, len(s.history))
	copy(result, s.history)
	return result
}

func (s *ActionStore) Replace(action model.UmbilicalAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = action
	s.history = []model.UmbilicalAction{action}
}
