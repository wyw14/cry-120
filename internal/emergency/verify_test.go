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

type blockingDriver struct {
	started chan struct{}
	release <-chan struct{}
}

func (d *blockingDriver) Drive(ctx context.Context, valve, position string, now time.Time) (manifold.DriveResult, error) {
	return manifold.DriveResult{Valve: valve, Position: position, Confirmed: true, CompletedAt: now}, nil
}

func (d *blockingDriver) Drain(ctx context.Context, valve string, now time.Time) (manifold.DriveResult, error) {
	select {
	case d.started <- struct{}{}:
	default:
	}
	select {
	case <-d.release:
		return manifold.DriveResult{Valve: valve, Position: "open", Confirmed: true, CompletedAt: now}, nil
	case <-ctx.Done():
		return manifold.DriveResult{}, ctx.Err()
	}
}

func TestFillCancellationWaitsForSafingWorkers(t *testing.T) {
	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	events, err := journal.Open(filepath.Join(t.TempDir(), "journal"))
	if err != nil {
		t.Fatal(err)
	}
	leases := manifold.NewLeaseManager()
	routes := manifold.NewRouteService(leases)
	sessions := propellant.NewSessionStore()
	start := propellant.NewStartService(sessions, routes, events)
	session, err := start.Start(context.Background(), model.Fuel, "fuel-arm", now)
	if err != nil {
		t.Fatal(err)
	}
	states := countdown.NewStateStore(120, now)
	controller := countdown.NewController(states, events)
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	safing := emergency.NewService(&blockingDriver{started: started, release: release}, routes, leases, interlock.NewHoldAggregate(), controller, events, emergency.NewStateStore())
	cancel := propellant.NewCancelService(sessions, safing, controller, events)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := cancel.Cancel(request.Context(), session.ID, time.Now())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if result.Phase != model.FillCancelled {
			http.Error(writer, "not cancelled", http.StatusConflict)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	responses := make(chan int, 1)
	go func() {
		response, callErr := http.Post(server.URL, "application/json", http.NoBody)
		if callErr != nil {
			responses <- 0
			return
		}
		response.Body.Close()
		responses <- response.StatusCode
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("safing worker did not start")
	}
	select {
	case status := <-responses:
		t.Fatalf("cancel returned before safing worker stopped: %d", status)
	case <-time.After(40 * time.Millisecond):
	}
	if controller.OperationsAllowed() {
		t.Fatal("countdown became operable while safing was active")
	}
	close(release)
	select {
	case status := <-responses:
		if status != http.StatusOK {
			t.Fatalf("cancel finished with %d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not finish after safing worker stopped")
	}
	if !controller.OperationsAllowed() {
		t.Fatal("countdown remained blocked after safing converged")
	}
}
