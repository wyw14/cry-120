package model

import "errors"

var ErrConflict = errors.New("resource conflict")
var ErrNotFound = errors.New("not found")
var ErrInvalidTransition = errors.New("invalid state transition")
var ErrActuatorTimeout = errors.New("actuator timeout")
var ErrOperationPending = errors.New("operation pending")

type Result struct {
	Accepted    bool        `json:"accepted"`
	Retryable   bool        `json:"retryable"`
	Code        string      `json:"code"`
	Message     string      `json:"message"`
	OperationID OperationID `json:"operation_id"`
}

func Accepted(operation OperationID, message string) Result {
	return Result{Accepted: true, Code: "accepted", Message: message, OperationID: operation}
}

func Rejected(operation OperationID, code, message string, retryable bool) Result {
	return Result{Accepted: false, Retryable: retryable, Code: code, Message: message, OperationID: operation}
}
