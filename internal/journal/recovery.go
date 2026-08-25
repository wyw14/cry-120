package journal

import (
	"context"
	"fmt"
	"sync"
)

type RecoveryStep interface {
	RecoveryName() string
	Recover(context.Context) error
}

type RecoveryCoordinator struct {
	mu        sync.Mutex
	flow      RecoveryStep
	sessions  RecoveryStep
	countdown RecoveryStep
	completed []string
}

func NewRecoveryCoordinator(flow, sessions, countdown RecoveryStep) *RecoveryCoordinator {
	return &RecoveryCoordinator{flow: flow, sessions: sessions, countdown: countdown}
}

func (c *RecoveryCoordinator) Recover(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.completed = c.completed[:0]
	steps := []RecoveryStep{c.flow, c.sessions, c.countdown}
	for _, step := range steps {
		if step == nil {
			return fmt.Errorf("recovery step missing")
		}
		if err := step.Recover(ctx); err != nil {
			return fmt.Errorf("recover %s: %w", step.RecoveryName(), err)
		}
		c.completed = append(c.completed, step.RecoveryName())
	}
	return nil
}

func (c *RecoveryCoordinator) Completed() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]string, len(c.completed))
	copy(result, c.completed)
	return result
}

func (c *RecoveryCoordinator) FlowReadyBeforeControllers() bool {
	items := c.Completed()
	return len(items) == 3 && items[0] == "manifold" && items[1] == "propellant" && items[2] == "countdown"
}
