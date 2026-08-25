package interlock

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type ReleaseService struct {
	holds   *HoldAggregate
	journal *journal.Store
}

func NewReleaseService(holds *HoldAggregate, store *journal.Store) *ReleaseService {
	return &ReleaseService{holds: holds, journal: store}
}

func (s *ReleaseService) Release(ctx context.Context, source, reason string, operation model.OperationID, now time.Time) (model.Result, error) {
	if !s.holds.Has(source, reason) {
		return model.Rejected(operation, "hold_not_found", "the requested hold is not active", false), model.ErrNotFound
	}
	revision := model.NewRevision()
	event, err := journal.NewEvent("hold.released", source, revision, map[string]string{"reason": reason, "operation_id": operation.String()}, now)
	if err != nil {
		return model.Result{}, err
	}
	if err := s.journal.Append(ctx, event); err != nil {
		return model.Rejected(operation, "journal_failed", err.Error(), true), err
	}
	if !s.holds.Release(source, reason) {
		return model.Rejected(operation, "release_conflict", "hold changed during release", true), model.ErrConflict
	}
	return model.Accepted(operation, "hold released"), nil
}

func (s *ReleaseService) Publish(source, reason string, now time.Time) model.Hold {
	return s.holds.Publish(source, reason, now)
}

func (s *ReleaseService) Active() []model.Hold {
	return s.holds.Active()
}

func (s *ReleaseService) Sources(reason string) []string {
	return s.holds.Sources(reason)
}
