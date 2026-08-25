package journal

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/wyw14/cry-120/internal/model"
)

type OperationRecord struct {
	OperationID model.OperationID `json:"operation_id"`
	Kind        string            `json:"kind"`
	Result      json.RawMessage   `json:"result"`
}

type OperationStore struct {
	mu  sync.Mutex
	dir string
}

func NewOperationStore(dir string) (*OperationStore, error) {
	path := filepath.Join(dir, "operations")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}
	return &OperationStore{dir: path}, nil
}

func (s *OperationStore) Commit(ctx context.Context, operation model.OperationID, kind string, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	record := OperationRecord{OperationID: operation, Kind: kind, Result: encoded}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(s.dir, operation.String()+".json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}

func (s *OperationStore) Lookup(operation model.OperationID, target any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := os.ReadFile(filepath.Join(s.dir, operation.String()+".json"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var record OperationRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return false, err
	}
	if err := json.Unmarshal(record.Result, target); err != nil {
		return false, err
	}
	return true, nil
}
