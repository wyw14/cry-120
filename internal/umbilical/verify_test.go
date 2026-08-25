package umbilical_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/umbilical"
)

func TestUmbilicalQuorumCountsDistinctControllers(t *testing.T) {
	now := time.Date(2026, 8, 25, 9, 30, 0, 0, time.UTC)
	actions := umbilical.NewActionStore(model.NewIdentity("generation"), now)
	current := actions.Current()
	receipts := journal.NewReceiptStore()
	service := umbilical.NewFeedbackService(actions, interlock.NewFeedbackStore(), receipts, interlock.NewQuorum("controller-a", "controller-b"))
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var input struct {
			Message string `json:"message"`
			Gateway string `json:"gateway"`
		}
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		receipt := model.Receipt{MessageID: model.Identity(input.Message), ControllerID: "controller-a", GatewayID: input.Gateway, ActionToken: current.Token, ReceivedAt: time.Now()}
		if _, err := service.RecordController(request.Context(), receipt); err != nil {
			http.Error(writer, err.Error(), http.StatusConflict)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	for _, body := range []string{`{"message":"primary-copy","gateway":"primary"}`, `{"message":"backup-copy","gateway":"backup"}`} {
		response, err := http.Post(server.URL, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusAccepted {
			t.Fatalf("receipt failed with %d", response.StatusCode)
		}
	}
	if service.QuorumReached() {
		t.Fatal("mirrored controller receipt counted as two safety channels")
	}
	missing := service.MissingControllers()
	if len(missing) != 1 || missing[0] != "controller-b" {
		t.Fatalf("unexpected missing controller set: %+v", missing)
	}
	if len(service.TransportReceipts()) != 2 {
		t.Fatal("transport audit trail did not retain both gateway messages")
	}
	if _, err := receipts.Add(context.Background(), model.Receipt{MessageID: "primary-copy", ControllerID: "controller-a", GatewayID: "retry", ActionToken: current.Token}); err != nil {
		t.Fatal(err)
	}
}
