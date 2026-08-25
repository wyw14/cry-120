package propellant

import (
	"sync"

	"github.com/wyw14/cry-120/internal/model"
)

type SessionStore struct {
	mu     sync.RWMutex
	items  map[model.Identity]model.FillSession
	byKind map[model.PropellantKind]model.Identity
}

func NewSessionStore() *SessionStore {
	return &SessionStore{items: make(map[model.Identity]model.FillSession), byKind: make(map[model.PropellantKind]model.Identity)}
}

func (s *SessionStore) Update(session model.FillSession) model.FillSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[session.ID] = session
	s.byKind[session.Kind] = session.ID
	return session
}

func (s *SessionStore) Get(id model.Identity) (model.FillSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.items[id]
	return session, ok
}

func (s *SessionStore) Current(kind model.PropellantKind) (model.FillSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byKind[kind]
	if !ok {
		return model.FillSession{}, false
	}
	session, ok := s.items[id]
	return session, ok
}

func (s *SessionStore) List() []model.FillSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.FillSession, 0, len(s.items))
	for _, session := range s.items {
		result = append(result, session)
	}
	return result
}

func (s *SessionStore) Replace(items []model.FillSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[model.Identity]model.FillSession, len(items))
	s.byKind = make(map[model.PropellantKind]model.Identity)
	for _, session := range items {
		s.items[session.ID] = session
		if current, ok := s.items[s.byKind[session.Kind]]; !ok || current.UpdatedAt.Before(session.UpdatedAt) {
			s.byKind[session.Kind] = session.ID
		}
	}
}
