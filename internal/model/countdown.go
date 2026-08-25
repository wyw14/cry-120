package model

import "time"

type CountdownPhase string

const (
	CountdownHeld    CountdownPhase = "held"
	CountdownRunning CountdownPhase = "running"
	CountdownReady   CountdownPhase = "ready"
)

type CountdownState struct {
	Generation     Identity       `json:"generation"`
	Phase          CountdownPhase `json:"phase"`
	TMinusSeconds  int            `json:"t_minus_seconds"`
	StableUntil    time.Time      `json:"stable_until"`
	PermitRevision Revision       `json:"permit_revision"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func NewCountdown(seconds int, now time.Time) CountdownState {
	return CountdownState{Generation: NewIdentity("countdown"), Phase: CountdownHeld, TMinusSeconds: seconds, UpdatedAt: now.UTC()}
}

func (s CountdownState) Clone() CountdownState {
	return s
}

func (s CountdownState) CanAdvance(now time.Time) bool {
	return s.Phase == CountdownRunning && !now.Before(s.StableUntil)
}

type Hold struct {
	Source    string    `json:"source"`
	Reason    string    `json:"reason"`
	Revision  Revision  `json:"revision"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

func (h Hold) Key() string {
	return h.Source + "\x00" + h.Reason
}
