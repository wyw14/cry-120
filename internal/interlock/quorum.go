package interlock

import "github.com/wyw14/cry-120/internal/model"

type Quorum struct {
	required map[string]struct{}
}

func NewQuorum(controllers ...string) *Quorum {
	required := make(map[string]struct{}, len(controllers))
	for _, controller := range controllers {
		required[controller] = struct{}{}
	}
	return &Quorum{required: required}
}

func (q *Quorum) Reached(receipts []model.Receipt) bool {
	present := make(map[string]struct{})
	for _, receipt := range receipts {
		if _, ok := q.required[receipt.ControllerID]; ok {
			present[receipt.ControllerID] = struct{}{}
		}
	}
	return len(q.required) > 0 && len(present) == len(q.required)
}

func (q *Quorum) Missing(receipts []model.Receipt) []string {
	present := make(map[string]struct{})
	for _, receipt := range receipts {
		present[receipt.ControllerID] = struct{}{}
	}
	missing := []string{}
	for controller := range q.required {
		if _, ok := present[controller]; !ok {
			missing = append(missing, controller)
		}
	}
	return missing
}
