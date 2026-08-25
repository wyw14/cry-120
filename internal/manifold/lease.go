package manifold

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Lease struct {
	Resource   string         `json:"resource"`
	Owner      model.Identity `json:"owner"`
	Token      model.Token    `json:"token"`
	ExpiresAt  time.Time      `json:"expires_at"`
	Generation uint64         `json:"generation"`
}

func (l Lease) Active(now time.Time) bool {
	return !l.Owner.Empty() && now.Before(l.ExpiresAt)
}

type LeaseManager struct {
	mu          sync.Mutex
	leases      map[string]Lease
	generations map[string]uint64
}

func NewLeaseManager() *LeaseManager {
	return &LeaseManager{leases: make(map[string]Lease), generations: make(map[string]uint64)}
}

func (m *LeaseManager) Acquire(ctx context.Context, resource string, owner model.Identity, duration time.Duration, now time.Time) (Lease, error) {
	if err := ctx.Err(); err != nil {
		return Lease{}, err
	}
	// The transfer manifold serves exactly one media path at a time, so the
	// free-check and the lease write must be a single atomic section. Holding
	// the fence across the actuator positioning delay guarantees that a racing
	// second request observes the established lease and gets a retryable
	// conflict instead of clobbering it and producing a second owner.
	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.leases[resource]
	if current.Active(now) && current.Owner != owner {
		return Lease{}, model.ErrConflict
	}
	time.Sleep(3 * time.Millisecond)
	m.generations[resource]++
	lease := Lease{Resource: resource, Owner: owner, Token: model.NewToken("fence"), ExpiresAt: now.Add(duration).UTC(), Generation: m.generations[resource]}
	m.leases[resource] = lease
	return lease, nil
}

func (m *LeaseManager) Release(resource string, owner model.Identity, token model.Token) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.leases[resource]
	if !ok || current.Owner != owner || current.Token != token {
		return false
	}
	delete(m.leases, resource)
	return true
}

func (m *LeaseManager) Current(resource string) (Lease, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.leases[resource]
	return current, ok
}

func (m *LeaseManager) All() []Lease {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Lease, 0, len(m.leases))
	for _, lease := range m.leases {
		result = append(result, lease)
	}
	return result
}

func (m *LeaseManager) Restore(items []Lease) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.leases = make(map[string]Lease, len(items))
	for _, lease := range items {
		m.leases[lease.Resource] = lease
		if lease.Generation > m.generations[lease.Resource] {
			m.generations[lease.Resource] = lease.Generation
		}
	}
}
