package interlock

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type PermitStore struct {
	mu      sync.RWMutex
	journal *journal.Store
	items   map[string]model.Permit
}

func NewPermitStore(store *journal.Store) *PermitStore {
	return &PermitStore{journal: store, items: make(map[string]model.Permit)}
}

func (s *PermitStore) Publish(ctx context.Context, permit model.Permit) error {
	s.mu.Lock()
	s.items[permit.Kind] = permit
	s.mu.Unlock()
	event, err := journal.NewEvent("permit.published", permit.Kind, permit.Revision, permit, time.Now())
	if err != nil {
		return err
	}
	if err := s.journal.Append(ctx, event); err != nil {
		return err
	}
	return nil
}

func (s *PermitStore) Current(kind string) (model.Permit, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	permit, ok := s.items[kind]
	return permit, ok
}

func (s *PermitStore) Revoke(kind string, revision model.Revision) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	permit, ok := s.items[kind]
	if !ok || permit.Revision != revision {
		return false
	}
	delete(s.items, kind)
	return true
}

func (s *PermitStore) List() []model.Permit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Permit, 0, len(s.items))
	for _, permit := range s.items {
		result = append(result, permit)
	}
	return result
}

func (s *PermitStore) Restore(items []model.Permit) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]model.Permit, len(items))
	for _, permit := range items {
		s.items[permit.Kind] = permit
	}
}
