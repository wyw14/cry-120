package interlock

type Status struct {
	Holds    int      `json:"holds"`
	Permits  int      `json:"permits"`
	Safe     bool     `json:"safe"`
	Failures []string `json:"failures"`
}

type StatusService struct {
	holds     *HoldAggregate
	permits   *PermitStore
	evaluator *Evaluator
}

func NewStatusService(holds *HoldAggregate, permits *PermitStore, evaluator *Evaluator) *StatusService {
	return &StatusService{holds: holds, permits: permits, evaluator: evaluator}
}

func (s *StatusService) Current() Status {
	last := s.evaluator.Last()
	return Status{Holds: len(s.holds.Active()), Permits: len(s.permits.List()), Safe: last.Safe, Failures: last.Failures}
}

func (s *StatusService) Permit(kind string) (any, bool) {
	permit, ok := s.permits.Current(kind)
	if !ok {
		return nil, false
	}
	return permit, true
}
