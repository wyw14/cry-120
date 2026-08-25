package countdown

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type ResumeResult struct {
	OperationID model.OperationID    `json:"operation_id"`
	State       model.CountdownState `json:"state"`
}

type ResumeService struct {
	states     *StateStore
	operations *journal.OperationStore
	events     *journal.Store
}

func NewResumeService(states *StateStore, operations *journal.OperationStore, events *journal.Store) *ResumeService {
	return &ResumeService{states: states, operations: operations, events: events}
}

func (s *ResumeService) Resume(ctx context.Context, operation model.OperationID, stableFor time.Duration, now time.Time) (ResumeResult, error) {
	var prior ResumeResult
	found, err := s.operations.Lookup(operation, &prior)
	if err != nil {
		return ResumeResult{}, err
	}
	if found {
		return prior, nil
	}
	current := s.states.Current()
	current.Generation = model.NewIdentity("countdown")
	current.Phase = model.CountdownRunning
	current.StableUntil = now.Add(stableFor).UTC()
	current.UpdatedAt = now.UTC()
	event, err := journal.NewEvent("countdown.resumed", current.Generation.String(), model.NewRevision(), map[string]string{"operation_id": operation.String()}, now)
	if err != nil {
		return ResumeResult{}, err
	}
	if err := s.events.Append(ctx, event); err != nil {
		return ResumeResult{}, err
	}
	result := ResumeResult{OperationID: operation, State: s.states.Update(current)}
	if err := s.operations.Commit(ctx, operation, "countdown.resume", result); err != nil {
		return ResumeResult{}, err
	}
	return result, nil
}

func (s *ResumeService) Lookup(operation model.OperationID) (ResumeResult, bool, error) {
	var result ResumeResult
	found, err := s.operations.Lookup(operation, &result)
	return result, found, err
}

func (s *ResumeService) ActiveGenerations() []model.Identity {
	return s.states.ActiveGenerations()
}
