package model

import "time"

type FillPhase string

const (
	FillPrepared   FillPhase = "prepared"
	FillChilldown  FillPhase = "chilldown"
	FillFast       FillPhase = "fast_fill"
	FillReplenish  FillPhase = "replenish"
	FillSecured    FillPhase = "secured"
	FillCancelling FillPhase = "cancelling"
	FillCancelled  FillPhase = "cancelled"
)

type PropellantKind string

const (
	Fuel     PropellantKind = "fuel"
	Oxidizer PropellantKind = "oxidizer"
)

type FillSession struct {
	ID         Identity       `json:"id"`
	Epoch      Identity       `json:"epoch"`
	Kind       PropellantKind `json:"kind"`
	Arm        string         `json:"arm"`
	Phase      FillPhase      `json:"phase"`
	RouteToken Token          `json:"route_token"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func NewFillSession(kind PropellantKind, arm string, now time.Time) FillSession {
	return FillSession{ID: NewIdentity("fill"), Epoch: NewIdentity("epoch"), Kind: kind, Arm: arm, Phase: FillPrepared, UpdatedAt: now.UTC()}
}

func (s FillSession) Terminal() bool {
	return s.Phase == FillSecured || s.Phase == FillCancelled
}

type Telemetry struct {
	Epoch         Identity  `json:"epoch"`
	Tank          string    `json:"tank"`
	Pressure      float64   `json:"pressure"`
	Temperature   float64   `json:"temperature"`
	Level         float64   `json:"level"`
	ValvePosition float64   `json:"valve_position"`
	ObservedAt    time.Time `json:"observed_at"`
}

func (t Telemetry) Valid() bool {
	return !t.Epoch.Empty() && t.Tank != "" && !t.ObservedAt.IsZero()
}
