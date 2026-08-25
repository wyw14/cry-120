package api

import (
	"net/http"
	"time"

	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/weather"
)

func (s *Server) ingestTelemetry(writer http.ResponseWriter, request *http.Request) {
	var reading model.Telemetry
	if err := decodeJSON(request, &reading); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if reading.ObservedAt.IsZero() {
		reading.ObservedAt = time.Now().UTC()
	}
	if err := s.deps.Telemetry.Ingest(request.Context(), reading, time.Now()); err != nil {
		writeJSON(writer, statusFor(err), map[string]string{"error": err.Error()})
		return
	}
	writeJSON(writer, http.StatusAccepted, reading)
}

func (s *Server) telemetryStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"readings": s.deps.Telemetry.Snapshot()})
}

func (s *Server) updateWeather(writer http.ResponseWriter, request *http.Request) {
	var reading weather.Reading
	if err := decodeJSON(request, &reading); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if reading.ValidUntil.IsZero() {
		reading.ValidUntil = time.Now().Add(2 * time.Minute).UTC()
	}
	writeJSON(writer, http.StatusAccepted, s.deps.Weather.Update(reading, time.Now()))
}

func (s *Server) weatherStatus(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, s.deps.Weather.Current(time.Now()))
}
