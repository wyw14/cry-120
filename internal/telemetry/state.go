package telemetry

import (
	"sync"

	"github.com/wyw14/cry-120/internal/model"
)

type StateStore struct {
	mu      sync.RWMutex
	latest  map[string]model.Telemetry
	byEpoch map[model.Identity][]model.Telemetry
}

func NewStateStore() *StateStore {
	return &StateStore{latest: make(map[string]model.Telemetry), byEpoch: make(map[model.Identity][]model.Telemetry)}
}

func (s *StateStore) Add(reading model.Telemetry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest[reading.Tank] = reading
	s.byEpoch[reading.Epoch] = append(s.byEpoch[reading.Epoch], reading)
}

func (s *StateStore) All() []model.Telemetry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Telemetry, 0, len(s.latest))
	for _, reading := range s.latest {
		result = append(result, reading)
	}
	return result
}
