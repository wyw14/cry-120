package countdown_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
)

func TestHoldReleaseRetryRecoversOriginalCountdownGeneration(t *testing.T) {
	now := time.Date(2026, 8, 25, 8, 40, 0, 0, time.UTC)
	dir := t.TempDir()
	events, err := journal.Open(filepath.Join(dir, "journal"))
	if err != nil {
		t.Fatal(err)
	}
	operations, err := journal.NewOperationStore(filepath.Join(dir, "journal"))
	if err != nil {
		t.Fatal(err)
	}
	states := countdown.NewStateStore(60, now)
	service := countdown.NewResumeService(states, operations, events)
	operation := model.NewOperationID()
	loseFirstResponse := true
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := service.Resume(request.Context(), operation, 18*time.Second, now)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusGatewayTimeout)
			return
		}
		if loseFirstResponse {
			loseFirstResponse = false
			http.Error(writer, "result log unavailable", http.StatusGatewayTimeout)
			return
		}
		json.NewEncoder(writer).Encode(result)
	}))
	defer server.Close()
	first, err := http.Post(server.URL, "application/json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	first.Body.Close()
	if first.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("expected first response loss, got %d", first.StatusCode)
	}
	committed := states.Current()
	second, err := http.Post(server.URL, "application/json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Body.Close()
	if second.StatusCode != http.StatusOK {
		t.Fatalf("retry failed with %d", second.StatusCode)
	}
	var recovered countdown.ResumeResult
	if err := json.NewDecoder(second.Body).Decode(&recovered); err != nil {
		t.Fatal(err)
	}
	if recovered.State.Generation != committed.Generation {
		t.Fatal("retry forked countdown generation")
	}
	if !recovered.State.StableUntil.Equal(committed.StableUntil) {
		t.Fatal("retry reset the original stability wait")
	}
	if len(states.ActiveGenerations()) != 1 {
		t.Fatalf("multiple active countdown generations: %+v", states.ActiveGenerations())
	}
}
