package journal

import (
	"context"
	"sync"

	"github.com/wyw14/cry-120/internal/model"
)

type ReceiptStore struct {
	mu         sync.RWMutex
	items      map[model.Token][]model.Receipt
	messageIDs map[model.Identity]struct{}
}

func NewReceiptStore() *ReceiptStore {
	return &ReceiptStore{items: make(map[model.Token][]model.Receipt), messageIDs: make(map[model.Identity]struct{})}
}

func (s *ReceiptStore) Add(ctx context.Context, receipt model.Receipt) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.messageIDs[receipt.MessageID]; ok {
		return false, nil
	}
	s.messageIDs[receipt.MessageID] = struct{}{}
	s.items[receipt.ActionToken] = append(s.items[receipt.ActionToken], receipt)
	return true, nil
}

func (s *ReceiptStore) ForAction(token model.Token) []model.Receipt {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.items[token]
	result := make([]model.Receipt, len(items))
	copy(result, items)
	return result
}
