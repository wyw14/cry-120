package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/propellant"
)

type IngestService struct {
	states   *StateStore
	sessions *propellant.SessionStore
	holds    *interlock.HoldAggregate
	events   *journal.Store
}

func NewIngestService(states *StateStore, sessions *propellant.SessionStore, holds *interlock.HoldAggregate, events *journal.Store) *IngestService {
	return &IngestService{states: states, sessions: sessions, holds: holds, events: events}
}

func (s *IngestService) Ingest(ctx context.Context, reading model.Telemetry, now time.Time) error {
	if !reading.Valid() {
		return fmt.Errorf("invalid telemetry")
	}
	matched := false
	var active model.FillSession
	for _, session := range s.sessions.List() {
		if session.Epoch == reading.Epoch && !session.Terminal() {
			matched = true
			active = session
			break
		}
	}
	if !matched {
		return model.ErrConflict
	}
	event, err := journal.NewEvent("telemetry.ingested", reading.Epoch.String(), model.NewRevision(), reading, now)
	if err != nil {
		return err
	}
	if err := s.events.Append(ctx, event); err != nil {
		return err
	}
	s.states.Add(reading)
	if active.Phase == model.FillChilldown && reading.Temperature < -150 {
		active.Phase = model.FillFast
		active.UpdatedAt = now.UTC()
		s.sessions.Update(active)
	}
	if active.Phase == model.FillFast && reading.Level >= 95 {
		active.Phase = model.FillReplenish
		active.UpdatedAt = now.UTC()
		s.sessions.Update(active)
	}
	if reading.Pressure > 5.5 || reading.Temperature > -100 {
		s.holds.Publish("telemetry:"+reading.Tank, "propellant envelope exceeded", now)
	} else {
		s.holds.Release("telemetry:"+reading.Tank, "propellant envelope exceeded")
	}
	return nil
}

func (s *IngestService) Snapshot() []model.Telemetry {
	return s.states.All()
}
