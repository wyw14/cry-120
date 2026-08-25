package propellant

import (
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/model"
)

type InterlockService struct {
	holds *interlock.HoldAggregate
}

func NewInterlockService(holds *interlock.HoldAggregate) *InterlockService {
	return &InterlockService{holds: holds}
}

func (s *InterlockService) LeakDetected(session model.Identity, now time.Time) model.Hold {
	reason := "区域不安全"
	return s.holds.Publish("propellant:"+session.String(), reason, now)
}

func (s *InterlockService) LeakCleared(session model.Identity) bool {
	return s.holds.Release("propellant:"+session.String(), "区域不安全")
}

func (s *InterlockService) ActiveFor(session model.Identity) bool {
	return s.holds.Has("propellant:"+session.String(), "区域不安全")
}

func (s *InterlockService) Reasons() []model.Hold {
	return s.holds.Active()
}
