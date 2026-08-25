package emergency_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/emergency"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/propellant"
)

func TestDrainFailurePreventsPadSafeResult(t *testing.T) {
	now := time.Date(2026, 8, 25, 9, 10, 0, 0, time.UTC)
	events, err := journal.Open(filepath.Join(t.TempDir(), "journal"))
	if err != nil {
		t.Fatal(err)
	}
	leases := manifold.NewLeaseManager()
	routes := manifold.NewRouteService(leases)
	sessions := propellant.NewSessionStore()
	session, err := propellant.NewStartService(sessions, routes, events).Start(context.Background(), model.Oxidizer, "lox-arm", now)
	if err != nil {
		t.Fatal(err)
	}
	driver := manifold.NewDriver()
	driver.FailValve("lox-arm-drain", model.ErrActuatorTimeout)
	holds := interlock.NewHoldAggregate()
	controller := countdown.NewController(countdown.NewStateStore(60, now), events)
	states := emergency.NewStateStore()
	service := emergency.NewService(driver, routes, leases, holds, controller, events, states)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if err := service.SafeSession(request.Context(), session, time.Now()); err != nil {
			http.Error(writer, err.Error(), http.StatusFailedDependency)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	response, err := http.Post(server.URL, "application/json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusFailedDependency {
		t.Fatalf("drain failure was reported safe, status=%d", response.StatusCode)
	}
	result, ok := states.Get(session.ID)
	if !ok || result.Safe {
		t.Fatalf("missing unsafe result: %+v", result)
	}
	if !holds.Has("emergency", "pad not safe") {
		t.Fatal("emergency hold was not retained")
	}
	if controller.OperationsAllowed() {
		t.Fatal("countdown became operable after drain failure")
	}
}
