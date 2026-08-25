package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-120/internal/model"
)

type startFillRequest struct {
	OperationID string               `json:"operation_id"`
	Kind        model.PropellantKind `json:"kind"`
	Arm         string               `json:"arm"`
}

func (s *Server) startFill(writer http.ResponseWriter, request *http.Request) {
	var input startFillRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, model.Rejected("", "invalid_json", err.Error(), false))
		return
	}
	operation := parseOperation(input.OperationID)
	result, session := s.deps.FillStart.StartResult(request.Context(), operation, input.Kind, input.Arm, time.Now())
	status := http.StatusAccepted
	if !result.Accepted {
		status = http.StatusConflict
	}
	writeJSON(writer, status, map[string]any{"accepted": result.Accepted, "retryable": result.Retryable, "code": result.Code, "message": result.Message, "operation_id": result.OperationID, "session": session})
}

type switchFillRequest struct {
	Arm string `json:"arm"`
}

func (s *Server) switchFill(writer http.ResponseWriter, request *http.Request) {
	var input switchFillRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	session, err := s.deps.FillSwitch.Switch(request.Context(), model.Identity(chi.URLParam(request, "session")), input.Arm, time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	path, _ := s.deps.FillSwitch.Route(input.Arm)
	writeJSON(writer, http.StatusOK, map[string]any{"session": session, "route": path})
}

func (s *Server) cancelFill(writer http.ResponseWriter, request *http.Request) {
	session, err := s.deps.FillCancel.Cancel(request.Context(), model.Identity(chi.URLParam(request, "session")), time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, session)
}

func (s *Server) raiseLeak(writer http.ResponseWriter, request *http.Request) {
	session := model.Identity(chi.URLParam(request, "session"))
	hold := s.deps.FillInterlock.LeakDetected(session, time.Now())
	writeJSON(writer, http.StatusCreated, map[string]any{"hold": hold, "active": s.deps.FillInterlock.ActiveFor(session)})
}

func (s *Server) clearLeak(writer http.ResponseWriter, request *http.Request) {
	session := model.Identity(chi.URLParam(request, "session"))
	if !s.deps.FillInterlock.LeakCleared(session) {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": "leak hold not found"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"active": s.deps.FillInterlock.ActiveFor(session)})
}
