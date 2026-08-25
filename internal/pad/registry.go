package pad

import (
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Vehicle struct {
	ID         model.Identity `json:"id"`
	Name       string         `json:"name"`
	Mission    string         `json:"mission"`
	PadID      string         `json:"pad_id"`
	AttachedAt time.Time      `json:"attached_at"`
}

type Registry struct {
	mu       sync.RWMutex
	vehicles map[string]Vehicle
}

func NewRegistry() *Registry {
	return &Registry{vehicles: make(map[string]Vehicle)}
}

func (r *Registry) Attach(padID, name, mission string, now time.Time) Vehicle {
	r.mu.Lock()
	defer r.mu.Unlock()
	vehicle := Vehicle{ID: model.NewIdentity("vehicle"), Name: name, Mission: mission, PadID: padID, AttachedAt: now.UTC()}
	r.vehicles[padID] = vehicle
	return vehicle
}

func (r *Registry) Current(padID string) (Vehicle, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	vehicle, ok := r.vehicles[padID]
	return vehicle, ok
}

func (r *Registry) List() []Vehicle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Vehicle, 0, len(r.vehicles))
	for _, vehicle := range r.vehicles {
		items = append(items, vehicle)
	}
	return items
}
