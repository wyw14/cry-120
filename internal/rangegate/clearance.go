package rangegate

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type ClearanceService struct {
	evidence *journal.EvidenceStore
	permits  *interlock.PermitStore
	holds    *countdown.HoldController
}

func NewClearanceService(evidence *journal.EvidenceStore, permits *interlock.PermitStore, holds *countdown.HoldController) *ClearanceService {
	return &ClearanceService{evidence: evidence, permits: permits, holds: holds}
}

func (s *ClearanceService) Submit(ctx context.Context, proof model.ClearanceProof, generation model.Identity) (model.Permit, error) {
	permit := model.Permit{Kind: "range-clear", Revision: proof.Revision, EvidenceID: proof.ID, Generation: generation, IssuedAt: time.Now().UTC()}
	if err := s.permits.Publish(ctx, permit); err != nil {
		return model.Permit{}, err
	}
	s.holds.Release("range", "range not clear")
	if err := s.evidence.SaveClearance(ctx, proof); err != nil {
		return model.Permit{}, err
	}
	return permit, nil
}

func (s *ClearanceService) EvidenceDurable(permit model.Permit) bool {
	return s.evidence.Exists(permit.EvidenceID)
}

func (s *ClearanceService) Current() (model.Permit, bool) {
	return s.permits.Current("range-clear")
}

func (s *ClearanceService) Revoke(permit model.Permit) bool {
	return s.permits.Revoke("range-clear", permit.Revision)
}
