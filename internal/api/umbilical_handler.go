package api

import (
	"net/http"
	"time"

	"github.com/wyw14/cry-120/internal/model"
)

func (s *Server) armUmbilical(writer http.ResponseWriter, request *http.Request) {
	action, err := s.deps.UmbilicalRetract.Arm(request.Context(), s.deps.Countdown.Current().Generation, time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, action)
}

type feedbackRequest struct {
	Kind         string `json:"kind"`
	Device       string `json:"device"`
	State        string `json:"state"`
	MessageID    string `json:"message_id"`
	ControllerID string `json:"controller_id"`
	GatewayID    string `json:"gateway_id"`
}

func (s *Server) umbilicalFeedback(writer http.ResponseWriter, request *http.Request) {
	var input feedbackRequest
	if err := decodeJSON(request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	action := s.deps.UmbilicalActions.Current()
	if input.Kind == "controller" {
		receipt := model.Receipt{MessageID: model.Identity(input.MessageID), ControllerID: input.ControllerID, GatewayID: input.GatewayID, ActionToken: action.Token, ReceivedAt: time.Now().UTC()}
		added, err := s.deps.UmbilicalFeedback.RecordController(request.Context(), receipt)
		if err != nil {
			writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusAccepted, map[string]any{"added": added, "quorum": s.deps.UmbilicalFeedback.QuorumReached()})
		return
	}
	feedback, err := s.deps.UmbilicalFeedback.RecordDevice(action.Token, input.Device, input.State, time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, feedback)
}

func (s *Server) retractUmbilical(writer http.ResponseWriter, request *http.Request) {
	action, err := s.deps.UmbilicalRetract.Start(request.Context(), s.deps.Countdown.Current().Generation, time.Now())
	if err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, action)
}
