package countdown

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Controller struct {
	mu      sync.Mutex
	states  *StateStore
	journal *journal.Store
	blocked bool
}

func NewController(states *StateStore, store *journal.Store) *Controller {
	return &Controller{states: states, journal: store, blocked: true}
}

func (c *Controller) Current() model.CountdownState {
	return c.states.Current()
}

func (c *Controller) Hold(ctx context.Context, revision model.Revision, reason string, now time.Time) (model.CountdownState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.states.Current()
	current.Phase = model.CountdownHeld
	current.PermitRevision = revision
	current.UpdatedAt = now.UTC()
	event, err := journal.NewEvent("countdown.held", current.Generation.String(), revision, map[string]string{"reason": reason}, now)
	if err != nil {
		return model.CountdownState{}, err
	}
	if err := c.journal.Append(ctx, event); err != nil {
		return model.CountdownState{}, err
	}
	c.blocked = true
	return c.states.Update(current), nil
}

func (c *Controller) AllowOperations(allowed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blocked = !allowed
}

func (c *Controller) OperationsAllowed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.blocked
}

func (c *Controller) Tick(now time.Time) model.CountdownState {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.states.Current()
	if current.CanAdvance(now) && current.TMinusSeconds > 0 {
		current.TMinusSeconds--
		current.UpdatedAt = now.UTC()
		if current.TMinusSeconds == 0 {
			current.Phase = model.CountdownReady
		}
		return c.states.Update(current)
	}
	return current
}
