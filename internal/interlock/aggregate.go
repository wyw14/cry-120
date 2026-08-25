package interlock

import (
	"sort"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type HoldAggregate struct {
	mu    sync.RWMutex
	holds map[string]model.Hold
}

func NewHoldAggregate() *HoldAggregate {
	return &HoldAggregate{holds: make(map[string]model.Hold)}
}

func (a *HoldAggregate) Publish(source, reason string, now time.Time) model.Hold {
	a.mu.Lock()
	defer a.mu.Unlock()
	hold := model.Hold{Source: source, Reason: reason, Revision: model.NewRevision(), Active: true, CreatedAt: now.UTC()}
	a.holds[hold.Key()] = hold
	return hold
}

func (a *HoldAggregate) Release(source, reason string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := model.Hold{Source: source, Reason: reason}.Key()
	if _, ok := a.holds[key]; !ok {
		return false
	}
	delete(a.holds, key)
	return true
}

func (a *HoldAggregate) Active() []model.Hold {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]model.Hold, 0, len(a.holds))
	for _, hold := range a.holds {
		result = append(result, hold)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Reason == result[j].Reason {
			return result[i].Source < result[j].Source
		}
		return result[i].Reason < result[j].Reason
	})
	return result
}

func (a *HoldAggregate) Sources(reason string) []string {
	items := a.Active()
	sources := []string{}
	for _, hold := range items {
		if hold.Reason == reason {
			sources = append(sources, hold.Source)
		}
	}
	return sources
}

func (a *HoldAggregate) Has(source, reason string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.holds[model.Hold{Source: source, Reason: reason}.Key()]
	return ok
}

func (a *HoldAggregate) Replace(items []model.Hold) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.holds = make(map[string]model.Hold, len(items))
	for _, hold := range items {
		if hold.Active {
			a.holds[hold.Key()] = hold
		}
	}
}
