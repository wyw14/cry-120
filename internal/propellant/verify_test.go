package propellant_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/propellant"
)

func TestConcurrentFillRequestsKeepTransferManifoldExclusive(t *testing.T) {
	events, err := journal.Open(filepath.Join(t.TempDir(), "journal"))
	if err != nil {
		t.Fatal(err)
	}
	leases := manifold.NewLeaseManager()
	routes := manifold.NewRouteService(leases)
	start := propellant.NewStartService(propellant.NewSessionStore(), routes, events)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var input struct {
			Kind model.PropellantKind `json:"kind"`
			Arm  string               `json:"arm"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := start.Start(request.Context(), input.Kind, input.Arm, time.Now())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusConflict)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	startGate := make(chan struct{})
	statuses := make(chan int, 2)
	var group sync.WaitGroup
	for _, body := range []string{`{"kind":"fuel","arm":"fuel-arm"}`, `{"kind":"oxidizer","arm":"lox-arm"}`} {
		group.Add(1)
		go func(payload string) {
			defer group.Done()
			<-startGate
			response, callErr := http.Post(server.URL, "application/json", strings.NewReader(payload))
			if callErr != nil {
				statuses <- 0
				return
			}
			response.Body.Close()
			statuses <- response.StatusCode
		}(body)
	}
	close(startGate)
	group.Wait()
	close(statuses)
	accepted := 0
	for status := range statuses {
		if status == http.StatusAccepted {
			accepted++
		}
	}
	if accepted != 1 {
		t.Fatalf("expected one accepted request, got %d", accepted)
	}
	items := leases.All()
	if len(items) != 1 || items[0].Resource != "transfer-manifold" {
		t.Fatalf("exclusive lease invariant failed: %+v", items)
	}
}
