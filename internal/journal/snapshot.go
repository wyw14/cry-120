package journal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type SnapshotStore struct {
	mu  sync.Mutex
	dir string
}

func NewSnapshotStore(dir string) (*SnapshotStore, error) {
	path := filepath.Join(dir, "snapshots")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}
	return &SnapshotStore{dir: path}, nil
}

func (s *SnapshotStore) Save(name string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	temporary := filepath.Join(s.dir, name+".tmp")
	final := filepath.Join(s.dir, name+".json")
	if err := os.WriteFile(temporary, encoded, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, final)
}

func (s *SnapshotStore) Load(name string, target any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	encoded, err := os.ReadFile(filepath.Join(s.dir, name+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		return false, err
	}
	return true, nil
}
