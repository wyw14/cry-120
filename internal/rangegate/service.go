package rangegate

import (
	"context"
	"errors"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type Request struct {
	OperationID model.OperationID `json:"operation_id"`
	Officer     string            `json:"officer"`
	Window      string            `json:"window"`
}

type Service struct {
	clearance *ClearanceService
}

func NewService(clearance *ClearanceService) *Service {
	return &Service{clearance: clearance}
}

func (s *Service) Confirm(ctx context.Context, request Request, generation model.Identity, now time.Time) (model.Result, model.Permit) {
	if request.OperationID == "" || request.Officer == "" || request.Window == "" {
		return model.Rejected(request.OperationID, "invalid_clearance", "operation, officer and window are required", false), model.Permit{}
	}
	proof := model.NewClearanceProof(request.OperationID, request.Officer, request.Window, now)
	permit, err := s.clearance.Submit(ctx, proof, generation)
	if err != nil {
		retryable := !errors.Is(err, context.Canceled)
		return model.Rejected(request.OperationID, "proof_write_failed", err.Error(), retryable), model.Permit{}
	}
	return model.Accepted(request.OperationID, "range clearance committed"), permit
}

func (s *Service) Status() map[string]any {
	permit, ok := s.clearance.Current()
	return map[string]any{"clear": ok, "permit": permit, "evidence_durable": ok && s.clearance.EvidenceDurable(permit)}
}
