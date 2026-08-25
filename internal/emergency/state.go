package emergency

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Step struct {
	Name        string    `json:"name"`
	Complete    bool      `json:"complete"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completed_at"`
}

type Result struct {
	OperationID model.OperationID `json:"operation_id"`
	SessionID   model.Identity    `json:"session_id"`
	Safe        bool              `json:"safe"`
	Steps       []Step            `json:"steps"`
	FinishedAt  time.Time         `json:"finished_at"`
}

type StateStore struct {
	mu    sync.RWMutex
	items map[model.Identity]Result
}

func NewStateStore() *StateStore {
	return &StateStore{items: make(map[model.Identity]Result)}
}

func (s *StateStore) Set(session model.Identity, result Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result.Steps = append([]Step{}, result.Steps...)
	s.items[session] = result
}

func (s *StateStore) Get(session model.Identity) (Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.items[session]
	result.Steps = append([]Step{}, result.Steps...)
	return result, ok
}

func (s *StateStore) List() []Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Result, 0, len(s.items))
	for _, item := range s.items {
		item.Steps = append([]Step{}, item.Steps...)
		result = append(result, item)
	}
	return result
}
