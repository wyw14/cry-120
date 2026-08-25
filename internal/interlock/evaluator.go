package interlock

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type SafetyResult struct {
	Safe        bool           `json:"safe"`
	Revision    model.Revision `json:"revision"`
	Failures    []string       `json:"failures"`
	EvaluatedAt time.Time      `json:"evaluated_at"`
}

type Evaluator struct {
	mu    sync.RWMutex
	holds *HoldAggregate
	last  SafetyResult
}

func NewEvaluator(holds *HoldAggregate) *Evaluator {
	return &Evaluator{holds: holds}
}

func (e *Evaluator) Evaluate(failures []string, now time.Time) SafetyResult {
	active := e.holds.Active()
	all := make([]string, 0, len(failures)+len(active))
	all = append(all, failures...)
	for _, hold := range active {
		all = append(all, hold.Source+":"+hold.Reason)
	}
	result := SafetyResult{Safe: len(all) == 0, Revision: model.NewRevision(), Failures: all, EvaluatedAt: now.UTC()}
	e.mu.Lock()
	e.last = result
	e.mu.Unlock()
	return result
}

func (e *Evaluator) Last() SafetyResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := e.last
	result.Failures = append([]string{}, result.Failures...)
	return result
}
