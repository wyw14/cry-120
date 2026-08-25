package manifold

import (
	"context"
	"sync"
	"time"
)

type DriveResult struct {
	Valve       string    `json:"valve"`
	Position    string    `json:"position"`
	Confirmed   bool      `json:"confirmed"`
	CompletedAt time.Time `json:"completed_at"`
}

type Driver struct {
	mu       sync.Mutex
	failures map[string]error
	results  []DriveResult
}

func NewDriver() *Driver {
	return &Driver{failures: make(map[string]error)}
}

func (d *Driver) FailValve(valve string, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.failures[valve] = err
}

func (d *Driver) Drive(ctx context.Context, valve, position string, now time.Time) (DriveResult, error) {
	if err := ctx.Err(); err != nil {
		return DriveResult{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.failures[valve]; err != nil {
		delete(d.failures, valve)
		return DriveResult{Valve: valve, Position: position, Confirmed: false, CompletedAt: now.UTC()}, err
	}
	result := DriveResult{Valve: valve, Position: position, Confirmed: true, CompletedAt: now.UTC()}
	d.results = append(d.results, result)
	return result, nil
}

func (d *Driver) History() []DriveResult {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make([]DriveResult, len(d.results))
	copy(result, d.results)
	return result
}

func (d *Driver) Drain(ctx context.Context, valve string, now time.Time) (DriveResult, error) {
	result, err := d.Drive(ctx, valve, "open", now)
	if err != nil {
		return result, nil
	}
	return result, nil
}
