package umbilical

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type FeedbackService struct {
	actions  *ActionStore
	feedback *interlock.FeedbackStore
	receipts *journal.ReceiptStore
	quorum   *interlock.Quorum
}

func NewFeedbackService(actions *ActionStore, feedback *interlock.FeedbackStore, receipts *journal.ReceiptStore, quorum *interlock.Quorum) *FeedbackService {
	return &FeedbackService{actions: actions, feedback: feedback, receipts: receipts, quorum: quorum}
}

func (s *FeedbackService) RecordDevice(token model.Token, device, state string, now time.Time) (interlock.Feedback, error) {
	action := s.actions.Current()
	if action.Token != token {
		return interlock.Feedback{}, model.ErrConflict
	}
	value := interlock.Feedback{ActionToken: token, Device: device, State: state, ObservedAt: now.UTC()}
	return s.feedback.Record(value), nil
}

func (s *FeedbackService) RecordController(ctx context.Context, receipt model.Receipt) (bool, error) {
	action := s.actions.Current()
	if receipt.ActionToken != action.Token {
		return false, model.ErrConflict
	}
	if receipt.MessageID.Empty() || receipt.ControllerID == "" || receipt.GatewayID == "" {
		return false, model.ErrInvalidTransition
	}
	return s.receipts.Add(ctx, receipt)
}

func (s *FeedbackService) QuorumReached() bool {
	action := s.actions.Current()
	return s.quorum.Reached(s.receipts.ForAction(action.Token))
}

func (s *FeedbackService) MissingControllers() []string {
	action := s.actions.Current()
	return s.quorum.Missing(s.receipts.ForAction(action.Token))
}

func (s *FeedbackService) TransportReceipts() []model.Receipt {
	action := s.actions.Current()
	return s.receipts.ForAction(action.Token)
}
