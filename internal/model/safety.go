package model

import "time"

type Permit struct {
	Kind       string    `json:"kind"`
	Revision   Revision  `json:"revision"`
	EvidenceID Identity  `json:"evidence_id"`
	Generation Identity  `json:"generation"`
	IssuedAt   time.Time `json:"issued_at"`
}

type ClearanceProof struct {
	ID           Identity    `json:"id"`
	OperationID  OperationID `json:"operation_id"`
	RangeOfficer string      `json:"range_officer"`
	Window       string      `json:"window"`
	Revision     Revision    `json:"revision"`
	SignedAt     time.Time   `json:"signed_at"`
}

func NewClearanceProof(operation OperationID, officer, window string, now time.Time) ClearanceProof {
	return ClearanceProof{ID: NewIdentity("proof"), OperationID: operation, RangeOfficer: officer, Window: window, Revision: NewRevision(), SignedAt: now.UTC()}
}

type UmbilicalState string

const (
	UmbilicalConnected  UmbilicalState = "connected"
	UmbilicalArmed      UmbilicalState = "release_armed"
	UmbilicalRetracting UmbilicalState = "retracting"
	UmbilicalRetracted  UmbilicalState = "retracted"
)

type UmbilicalAction struct {
	Token         Token          `json:"token"`
	Generation    Identity       `json:"generation"`
	State         UmbilicalState `json:"state"`
	ReadyRevision Revision       `json:"ready_revision"`
	Controllers   []string       `json:"controllers"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type Receipt struct {
	MessageID    Identity  `json:"message_id"`
	ControllerID string    `json:"controller_id"`
	GatewayID    string    `json:"gateway_id"`
	ActionToken  Token     `json:"action_token"`
	ReceivedAt   time.Time `json:"received_at"`
}
