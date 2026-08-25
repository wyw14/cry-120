package interlock

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Feedback struct {
	ActionToken model.Token    `json:"action_token"`
	Device      string         `json:"device"`
	State       string         `json:"state"`
	Revision    model.Revision `json:"revision"`
	ObservedAt  time.Time      `json:"observed_at"`
}

type FeedbackStore struct {
	mu    sync.RWMutex
	items map[model.Token]map[string]Feedback
}

func NewFeedbackStore() *FeedbackStore {
	return &FeedbackStore{items: make(map[model.Token]map[string]Feedback)}
}

func (s *FeedbackStore) Record(value Feedback) Feedback {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value.Revision == "" {
		value.Revision = model.NewRevision()
	}
	if value.ObservedAt.IsZero() {
		value.ObservedAt = time.Now().UTC()
	}
	if s.items[value.ActionToken] == nil {
		s.items[value.ActionToken] = make(map[string]Feedback)
	}
	s.items[value.ActionToken][value.Device] = value
	return value
}

func (s *FeedbackStore) Complete(token model.Token, required ...string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.items[token]
	for _, device := range required {
		feedback, ok := items[device]
		if !ok || feedback.State != "ready" {
			return false
		}
	}
	return len(required) > 0
}

func (s *FeedbackStore) ForAction(token model.Token) []Feedback {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Feedback{}
	for _, value := range s.items[token] {
		result = append(result, value)
	}
	return result
}

func (s *FeedbackStore) Reset(token model.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, token)
}
