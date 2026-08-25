package rangegate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/rangegate"
)

func TestRangeClearanceWaitsForDurableProof(t *testing.T) {
	now := time.Date(2026, 8, 25, 8, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	events, err := journal.Open(filepath.Join(dir, "journal"))
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := journal.NewEvidenceStore(filepath.Join(dir, "journal"))
	if err != nil {
		t.Fatal(err)
	}
	permits := interlock.NewPermitStore(events)
	states := countdown.NewStateStore(30, now)
	controller := countdown.NewController(states, events)
	holds := countdown.NewHoldController(controller)
	hold := model.Hold{Source: "range", Reason: "range not clear", Revision: model.NewRevision(), Active: true, CreatedAt: now}
	if err := holds.Apply(context.Background(), hold, now); err != nil {
		t.Fatal(err)
	}
	service := rangegate.NewService(rangegate.NewClearanceService(evidence, permits, holds))
	proofDir := filepath.Join(dir, "journal", "proofs")
	if err := os.RemoveAll(proofDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proofDir, []byte("occupied"), 0o644); err != nil {
		t.Fatal(err)
	}
	var proofID model.Identity
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		input := rangegate.Request{OperationID: model.NewOperationID(), Officer: "range-one", Window: "T-30"}
		result, permit := service.Confirm(request.Context(), input, states.Current().Generation, now)
		proofID = permit.EvidenceID
		writer.Header().Set("Content-Type", "application/json")
		if !result.Accepted {
			writer.WriteHeader(http.StatusInsufficientStorage)
		}
		json.NewEncoder(writer).Encode(result)
	}))
	defer server.Close()
	response, err := http.Post(server.URL, "application/json", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusInsufficientStorage {
		t.Fatalf("expected storage failure, got %d", response.StatusCode)
	}
	if _, ok := permits.Current("range-clear"); ok {
		t.Fatal("range permit became public without durable proof")
	}
	if controller.OperationsAllowed() {
		t.Fatal("countdown hold was released after proof write failure")
	}
	if proofID != "" && evidence.Exists(proofID) {
		t.Fatal("failed proof unexpectedly exists")
	}
}
