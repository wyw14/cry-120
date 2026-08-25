package manifold

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type ValveCommand struct {
	ID       model.Identity `json:"id"`
	Valve    string         `json:"valve"`
	Position string         `json:"position"`
	Owner    model.Identity `json:"owner"`
	Token    model.Token    `json:"token"`
	IssuedAt time.Time      `json:"issued_at"`
}

type Commander struct {
	mu       sync.Mutex
	leases   *LeaseManager
	commands []ValveCommand
}

func NewCommander(leases *LeaseManager) *Commander {
	return &Commander{leases: leases}
}

func (c *Commander) Issue(ctx context.Context, resource, valve, position string, owner model.Identity, token model.Token, now time.Time) (ValveCommand, error) {
	if err := ctx.Err(); err != nil {
		return ValveCommand{}, err
	}
	lease, ok := c.leases.Current(resource)
	if !ok || lease.Owner != owner || lease.Token != token {
		return ValveCommand{}, model.ErrConflict
	}
	command := ValveCommand{ID: model.NewIdentity("command"), Valve: valve, Position: position, Owner: owner, Token: token, IssuedAt: now.UTC()}
	c.mu.Lock()
	c.commands = append(c.commands, command)
	c.mu.Unlock()
	return command, nil
}

func (c *Commander) History() []ValveCommand {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]ValveCommand, len(c.commands))
	copy(result, c.commands)
	return result
}
