package pad

import (
	"sync"
	"time"
)

type Condition struct {
	PadID        string    `json:"pad_id"`
	Evacuated    bool      `json:"evacuated"`
	GroundPower  bool      `json:"ground_power"`
	AccessClosed bool      `json:"access_closed"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Conditions struct {
	mu    sync.RWMutex
	items map[string]Condition
}

func NewConditions() *Conditions {
	return &Conditions{items: make(map[string]Condition)}
}

func (c *Conditions) Set(value Condition) Condition {
	c.mu.Lock()
	defer c.mu.Unlock()
	value.UpdatedAt = value.UpdatedAt.UTC()
	c.items[value.PadID] = value
	return value
}

func (c *Conditions) Get(padID string) (Condition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.items[padID]
	return value, ok
}

func (c *Conditions) SafeForOperations(padID string) bool {
	value, ok := c.Get(padID)
	return ok && value.Evacuated && value.GroundPower && value.AccessClosed
}
