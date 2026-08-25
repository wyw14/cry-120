package api

import (
	"net/http"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

type holdRequest struct {
	Source      string `json:"source"`
	Reason      string `json:"reason"`
	OperationID string `json:"operation_id"`
}

func (s *Server) publishHold(writer http.ResponseWriter, request *http.Request) {
	var input holdRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.Source == "" || input.Reason == "" {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "source and reason are required"})
		return
	}
	hold := s.deps.Holds.Publish(input.Source, input.Reason, time.Now())
	writeJSON(writer, http.StatusCreated, hold)
}

func (s *Server) releaseHold(writer http.ResponseWriter, request *http.Request) {
	var input holdRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	result, err := s.deps.Holds.Release(request.Context(), input.Source, input.Reason, parseOperation(input.OperationID), time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), result)
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

type resumeRequest struct {
	OperationID   string `json:"operation_id"`
	StableSeconds int    `json:"stable_seconds"`
}

func (s *Server) resumeCountdown(writer http.ResponseWriter, request *http.Request) {
	var input resumeRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.StableSeconds <= 0 {
		input.StableSeconds = 20
	}
	result, err := s.deps.Resume.Resume(request.Context(), parseOperation(input.OperationID), time.Duration(input.StableSeconds)*time.Second, time.Now())
	if err != nil {
		writeJSON(writer, http.StatusGatewayTimeout, model.Rejected(parseOperation(input.OperationID), "result_unavailable", err.Error(), true))
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

type recycleRequest struct {
	Seconds int `json:"seconds"`
}

func (s *Server) recycleCountdown(writer http.ResponseWriter, request *http.Request) {
	var input recycleRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.Seconds < 90 {
		input.Seconds = 300
	}
	state, err := s.deps.Recycler.Recycle(request.Context(), input.Seconds, time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, state)
}

func (s *Server) tickCountdown(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, s.deps.Countdown.Tick(time.Now()))
}
