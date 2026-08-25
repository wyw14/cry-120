package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/emergency"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/pad"
	"github.com/wyw14/cry-120/internal/propellant"
	"github.com/wyw14/cry-120/internal/rangegate"
	"github.com/wyw14/cry-120/internal/telemetry"
	"github.com/wyw14/cry-120/internal/umbilical"
	"github.com/wyw14/cry-120/internal/weather"
	"github.com/wyw14/cry-120/internal/web"
)

type Dependencies struct {
	Countdown         *countdown.Controller
	Resume            *countdown.ResumeService
	Recycler          *countdown.Recycler
	FillStart         *propellant.StartService
	FillSwitch        *propellant.SwitchService
	FillCancel        *propellant.CancelService
	FillStatus        *propellant.StatusService
	FillInterlock     *propellant.InterlockService
	UmbilicalRetract  *umbilical.Retractor
	UmbilicalFeedback *umbilical.FeedbackService
	UmbilicalActions  *umbilical.ActionStore
	Interlock         *interlock.StatusService
	Holds             *interlock.ReleaseService
	Range             *rangegate.Service
	Telemetry         *telemetry.IngestService
	Weather           *weather.Service
	Emergency         *emergency.StateStore
	Actuators         *manifold.Driver
	Pads              *pad.Registry
	PadConditions     *pad.Conditions
	Pages             *web.Pages
}

type Server struct {
	deps      Dependencies
	router    chi.Router
	startedAt time.Time
}

func NewServer(deps Dependencies) *Server {
	server := &Server{deps: deps, router: chi.NewRouter(), startedAt: time.Now().UTC()}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() {
	s.router.Get("/healthz", s.health)
	s.router.Get("/api/countdown", s.countdownStatus)
	s.router.Post("/api/countdown/resume", s.resumeCountdown)
	s.router.Post("/api/countdown/recycle", s.recycleCountdown)
	s.router.Post("/api/countdown/tick", s.tickCountdown)
	s.router.Get("/api/propellant", s.propellantStatus)
	s.router.Post("/api/propellant/start", s.startFill)
	s.router.Post("/api/propellant/{session}/switch", s.switchFill)
	s.router.Post("/api/propellant/{session}/cancel", s.cancelFill)
	s.router.Post("/api/propellant/{session}/leak", s.raiseLeak)
	s.router.Delete("/api/propellant/{session}/leak", s.clearLeak)
	s.router.Get("/api/umbilical", s.umbilicalStatus)
	s.router.Post("/api/umbilical/arm", s.armUmbilical)
	s.router.Post("/api/umbilical/feedback", s.umbilicalFeedback)
	s.router.Post("/api/umbilical/retract", s.retractUmbilical)
	s.router.Get("/api/interlocks", s.interlockStatus)
	s.router.Get("/api/holds", s.holdStatus)
	s.router.Post("/api/holds", s.publishHold)
	s.router.Delete("/api/holds", s.releaseHold)
	s.router.Post("/api/range/clear", s.clearRange)
	s.router.Get("/api/range", s.rangeStatus)
	s.router.Post("/api/telemetry", s.ingestTelemetry)
	s.router.Get("/api/telemetry", s.telemetryStatus)
	s.router.Post("/api/weather", s.updateWeather)
	s.router.Get("/api/weather", s.weatherStatus)
	s.router.Post("/api/interlocks/actuators/{valve}/fault", s.recordActuatorFault)
	for _, path := range s.deps.Pages.Paths() {
		path := path
		s.router.Get(path, func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := s.deps.Pages.Render(path, writer); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		})
	}
}
