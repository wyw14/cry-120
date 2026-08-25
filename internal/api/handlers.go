package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	json.NewEncoder(writer).Encode(value)
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, request.Body, 1024*1024))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, model.ErrConflict), errors.Is(err, model.ErrOperationPending):
		return http.StatusConflict
	case errors.Is(err, model.ErrInvalidTransition):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "started_at": s.startedAt, "uptime_seconds": int(time.Since(s.startedAt).Seconds())})
}

func (s *Server) countdownStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"state": s.deps.Countdown.Current(), "operations_allowed": s.deps.Countdown.OperationsAllowed(), "active_generations": s.deps.Resume.ActiveGenerations()})
}

func (s *Server) propellantStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"flow": s.deps.FillStatus.Current(), "holds": s.deps.FillInterlock.Reasons()})
}

func (s *Server) umbilicalStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"action": s.deps.UmbilicalActions.Current(), "quorum": s.deps.UmbilicalFeedback.QuorumReached(), "missing_controllers": s.deps.UmbilicalFeedback.MissingControllers(), "receipts": s.deps.UmbilicalFeedback.TransportReceipts()})
}

func (s *Server) interlockStatus(writer http.ResponseWriter, request *http.Request) {
	permit, _ := s.deps.Interlock.Permit("range-clear")
	writeJSON(writer, http.StatusOK, map[string]any{"status": s.deps.Interlock.Current(), "range_permit": permit, "safing_results": s.deps.Emergency.List()})
}

func (s *Server) holdStatus(writer http.ResponseWriter, request *http.Request) {
	holds := s.deps.Holds.Active()
	owners := make(map[string][]string)
	for _, hold := range holds {
		owners[hold.Reason] = s.deps.Holds.Sources(hold.Reason)
	}
	writeJSON(writer, http.StatusOK, map[string]any{"holds": holds, "owners": owners})
}

func parseOperation(value string) model.OperationID {
	if value == "" {
		return model.NewOperationID()
	}
	return model.OperationID(value)
}
