package emergency

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
)

type ValveDriver interface {
	Drive(context.Context, string, string, time.Time) (manifold.DriveResult, error)
	Drain(context.Context, string, time.Time) (manifold.DriveResult, error)
}

type Service struct {
	driver    ValveDriver
	routes    *manifold.RouteService
	leases    *manifold.LeaseManager
	holds     *interlock.HoldAggregate
	countdown *countdown.Controller
	events    *journal.Store
	states    *StateStore
}

func NewService(driver ValveDriver, routes *manifold.RouteService, leases *manifold.LeaseManager, holds *interlock.HoldAggregate, controller *countdown.Controller, events *journal.Store, states *StateStore) *Service {
	return &Service{driver: driver, routes: routes, leases: leases, holds: holds, countdown: controller, events: events, states: states}
}

func (s *Service) SafeSession(ctx context.Context, session model.FillSession, now time.Time) error {
	result := Result{OperationID: model.NewOperationID(), SessionID: session.ID}
	steps := []struct {
		name string
		run  func() error
	}{
		{name: "cutoff", run: func() error {
			_, err := s.driver.Drive(ctx, session.Arm+"-isolation", "closed", now)
			return err
		}},
		{name: "drain", run: func() error {
			_, err := s.driver.Drain(ctx, session.Arm+"-drain", time.Now())
			return err
		}},
		{name: "route-release", run: func() error {
			if !s.routes.Close(session.Arm, session.ID) {
				return model.ErrConflict
			}
			if !s.leases.Release("transfer-manifold", session.ID, session.RouteToken) {
				return model.ErrConflict
			}
			return nil
		}},
	}
	for _, action := range steps {
		err := action.run()
		step := Step{Name: action.name, Complete: err == nil, CompletedAt: time.Now().UTC()}
		if err != nil {
			step.Error = err.Error()
			result.Steps = append(result.Steps, step)
			result.Safe = false
			result.FinishedAt = time.Now().UTC()
			s.states.Set(session.ID, result)
			s.holds.Publish("emergency", "pad not safe", time.Now())
			s.countdown.AllowOperations(false)
			return fmt.Errorf("safing %s: %w", action.name, err)
		}
		result.Steps = append(result.Steps, step)
	}
	event, err := journal.NewEvent("safing.completed", session.ID.String(), model.NewRevision(), result.Steps, time.Now())
	if err != nil {
		return err
	}
	if err := s.events.Append(ctx, event); err != nil {
		return err
	}
	result.Safe = true
	result.FinishedAt = time.Now().UTC()
	s.states.Set(session.ID, result)
	s.holds.Release("emergency", "pad not safe")
	return nil
}
