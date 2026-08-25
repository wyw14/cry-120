package countdown

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type HoldController struct {
	mu         sync.Mutex
	controller *Controller
	holds      map[string]model.Hold
}

func NewHoldController(controller *Controller) *HoldController {
	return &HoldController{controller: controller, holds: make(map[string]model.Hold)}
}

func (h *HoldController) Apply(ctx context.Context, hold model.Hold, now time.Time) error {
	h.mu.Lock()
	h.holds[hold.Key()] = hold
	h.mu.Unlock()
	_, err := h.controller.Hold(ctx, hold.Revision, hold.Reason, now)
	return err
}

func (h *HoldController) Release(source, reason string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := model.Hold{Source: source, Reason: reason}.Key()
	if _, ok := h.holds[key]; !ok {
		return false
	}
	delete(h.holds, key)
	if len(h.holds) == 0 {
		h.controller.AllowOperations(true)
	}
	return true
}

func (h *HoldController) Replace(items []model.Hold) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.holds = make(map[string]model.Hold, len(items))
	for _, hold := range items {
		if hold.Active {
			h.holds[hold.Key()] = hold
		}
	}
	h.controller.AllowOperations(len(h.holds) == 0)
}

func (h *HoldController) Active() []model.Hold {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]model.Hold, 0, len(h.holds))
	for _, hold := range h.holds {
		result = append(result, hold)
	}
	return result
}
