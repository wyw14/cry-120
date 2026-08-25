package journal

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wyw14/cry-120/internal/model"
)

type EvidenceStore struct {
	dir string
}

func NewEvidenceStore(dir string) (*EvidenceStore, error) {
	path := filepath.Join(dir, "proofs")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}
	return &EvidenceStore{dir: path}, nil
}

func (s *EvidenceStore) SaveClearance(ctx context.Context, proof model.ClearanceProof) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return err
	}
	temporary := filepath.Join(s.dir, proof.ID.String()+".tmp")
	final := filepath.Join(s.dir, proof.ID.String()+".json")
	if err := os.WriteFile(temporary, encoded, 0o644); err != nil {
		return err
	}
	file, err := os.OpenFile(temporary, os.O_RDWR, 0o644)
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
	return os.Rename(temporary, final)
}

func (s *EvidenceStore) Exists(id model.Identity) bool {
	_, err := os.Stat(filepath.Join(s.dir, id.String()+".json"))
	return err == nil
}
