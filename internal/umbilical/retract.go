package umbilical

import (
	"context"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

type Retractor struct {
	actions  *ActionStore
	feedback *interlock.FeedbackStore
	events   *journal.Store
}

func NewRetractor(actions *ActionStore, feedback *interlock.FeedbackStore, events *journal.Store) *Retractor {
	return &Retractor{actions: actions, feedback: feedback, events: events}
}

func (r *Retractor) Arm(ctx context.Context, generation model.Identity, now time.Time) (model.UmbilicalAction, error) {
	previous := r.actions.Current()
	if previous.Token != "" {
		r.feedback.Reset(previous.Token)
	}
	action := model.UmbilicalAction{Token: model.NewToken("umbilical"), Generation: generation, State: model.UmbilicalArmed, UpdatedAt: now.UTC()}
	event, err := journal.NewEvent("umbilical.armed", action.Token.String(), model.NewRevision(), action, now)
	if err != nil {
		return model.UmbilicalAction{}, err
	}
	if err := r.events.Append(ctx, event); err != nil {
		return model.UmbilicalAction{}, err
	}
	return r.actions.Update(action), nil
}

func (r *Retractor) Start(ctx context.Context, generation model.Identity, now time.Time) (model.UmbilicalAction, error) {
	action := r.actions.Current()
	if action.Generation != generation || action.State != model.UmbilicalArmed {
		return model.UmbilicalAction{}, model.ErrInvalidTransition
	}
	if !r.feedback.Complete(action.Token, "connector-unlocked", "arm-stable") {
		return model.UmbilicalAction{}, model.ErrOperationPending
	}
	action.State = model.UmbilicalRetracting
	action.ReadyRevision = model.NewRevision()
	action.UpdatedAt = now.UTC()
	event, err := journal.NewEvent("umbilical.retracting", action.Token.String(), action.ReadyRevision, action, now)
	if err != nil {
		return model.UmbilicalAction{}, err
	}
	if err := r.events.Append(ctx, event); err != nil {
		return model.UmbilicalAction{}, err
	}
	return r.actions.Update(action), nil
}

func (r *Retractor) Complete(now time.Time) (model.UmbilicalAction, error) {
	action := r.actions.Current()
	if action.State != model.UmbilicalRetracting {
		return model.UmbilicalAction{}, model.ErrInvalidTransition
	}
	action.State = model.UmbilicalRetracted
	action.UpdatedAt = now.UTC()
	return r.actions.Update(action), nil
}
