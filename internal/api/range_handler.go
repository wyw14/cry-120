package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/rangegate"
)

func (s *Server) clearRange(writer http.ResponseWriter, request *http.Request) {
	var input rangegate.Request
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	result, permit := s.deps.Range.Confirm(request.Context(), input, s.deps.Countdown.Current().Generation, time.Now())
	if !result.Accepted {
		writeJSON(writer, http.StatusInsufficientStorage, result)
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{"result": result, "permit": permit})
}

func (s *Server) rangeStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, s.deps.Range.Status())
}

func (s *Server) recordActuatorFault(writer http.ResponseWriter, request *http.Request) {
	valve := chi.URLParam(request, "valve")
	if valve == "" {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "valve is required"})
		return
	}
	s.deps.Actuators.FailValve(valve, model.ErrActuatorTimeout)
	writeJSON(writer, http.StatusAccepted, map[string]string{"valve": valve, "status": "fault recorded"})
}
