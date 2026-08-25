package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wyw14/cry-120/internal/api"
	"github.com/wyw14/cry-120/internal/countdown"
	"github.com/wyw14/cry-120/internal/emergency"
	"github.com/wyw14/cry-120/internal/interlock"
	"github.com/wyw14/cry-120/internal/journal"
	"github.com/wyw14/cry-120/internal/manifold"
	"github.com/wyw14/cry-120/internal/model"
	"github.com/wyw14/cry-120/internal/pad"
	"github.com/wyw14/cry-120/internal/propellant"
	"github.com/wyw14/cry-120/internal/rangegate"
	"github.com/wyw14/cry-120/internal/telemetry"
	"github.com/wyw14/cry-120/internal/umbilical"
	"github.com/wyw14/cry-120/internal/weather"
	"github.com/wyw14/cry-120/internal/web"
)

type application struct {
	handler           http.Handler
	recovery          *journal.RecoveryCoordinator
	flowRecovery      *manifold.Recovery
	fillRecovery      *propellant.Recovery
	countdownRecovery *countdown.Recovery
	interlockRecovery *interlock.Recovery
}

func assemble(runtimeDir string, now time.Time) (*application, error) {
	if runtimeDir == "" {
		runtimeDir = "runtime"
	}
	absolute, err := filepath.Abs(runtimeDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0o755); err != nil {
		return nil, err
	}
	events, err := journal.Open(filepath.Join(absolute, "journal"))
	if err != nil {
		return nil, err
	}
	evidence, err := journal.NewEvidenceStore(filepath.Join(absolute, "journal"))
	if err != nil {
		return nil, err
	}
	operations, err := journal.NewOperationStore(filepath.Join(absolute, "journal"))
	if err != nil {
		return nil, err
	}
	snapshots, err := journal.NewSnapshotStore(filepath.Join(absolute, "journal"))
	if err != nil {
		return nil, err
	}
	leases := manifold.NewLeaseManager()
	routes := manifold.NewRouteService(leases)
	commander := manifold.NewCommander(leases)
	driver := manifold.NewDriver()
	flowRecovery := manifold.NewRecovery(snapshots, leases)
	holdAggregate := interlock.NewHoldAggregate()
	permitStore := interlock.NewPermitStore(events)
	feedbackStore := interlock.NewFeedbackStore()
	evaluator := interlock.NewEvaluator(holdAggregate)
	interlockStatus := interlock.NewStatusService(holdAggregate, permitStore, evaluator)
	releaseService := interlock.NewReleaseService(holdAggregate, events)
	interlockRecovery := interlock.NewRecovery(snapshots, holdAggregate, permitStore)
	countdownStates := countdown.NewStateStore(300, now)
	countdownController := countdown.NewController(countdownStates, events)
	countdownHolds := countdown.NewHoldController(countdownController)
	resumeService := countdown.NewResumeService(countdownStates, operations, events)
	recycler := countdown.NewRecycler(countdownStates, events)
	countdownRecovery := countdown.NewRecovery(snapshots, countdownStates)
	sessions := propellant.NewSessionStore()
	startService := propellant.NewStartService(sessions, routes, events)
	switchService := propellant.NewSwitchService(sessions, routes, events)
	fillStatus := propellant.NewStatusService(sessions, routes, leases)
	propellantInterlock := propellant.NewInterlockService(holdAggregate)
	fillRecovery := propellant.NewRecovery(snapshots, sessions, flowRecovery, commander)
	emergencyStates := emergency.NewStateStore()
	emergencyService := emergency.NewService(driver, routes, leases, holdAggregate, countdownController, events, emergencyStates)
	cancelService := propellant.NewCancelService(sessions, emergencyService, countdownController, events)
	receipts := journal.NewReceiptStore()
	umbilicalActions := umbilical.NewActionStore(countdownStates.Current().Generation, now)
	umbilicalFeedback := umbilical.NewFeedbackService(umbilicalActions, feedbackStore, receipts, interlock.NewQuorum("controller-a", "controller-b"))
	retractor := umbilical.NewRetractor(umbilicalActions, feedbackStore, events)
	clearance := rangegate.NewClearanceService(evidence, permitStore, countdownHolds)
	rangeService := rangegate.NewService(clearance)
	weatherHolds := weather.NewHoldService(holdAggregate)
	weatherService := weather.NewService(weatherHolds)
	telemetryStates := telemetry.NewStateStore()
	telemetryService := telemetry.NewIngestService(telemetryStates, sessions, holdAggregate, events)
	padRegistry := pad.NewRegistry()
	padRegistry.Attach("LC-1", "LaunchGuard Demonstrator", "qualification", now)
	padConditions := pad.NewConditions()
	padConditions.Set(pad.Condition{PadID: "LC-1", Evacuated: true, GroundPower: true, AccessClosed: true, UpdatedAt: now})
	countdownHolds.Apply(context.Background(), releaseService.Publish("range", "range not clear", now), now)
	propellantInterlock.Reasons()
	weatherService.Update(weather.Reading{WindMetersPerSecond: 4, LightningDistanceKM: 80, ValidUntil: now.Add(5 * time.Minute)}, now)
	evaluator.Evaluate(nil, now)
	pages := web.NewPages()
	server := api.NewServer(api.Dependencies{
		Countdown:         countdownController,
		Resume:            resumeService,
		Recycler:          recycler,
		FillStart:         startService,
		FillSwitch:        switchService,
		FillCancel:        cancelService,
		FillStatus:        fillStatus,
		FillInterlock:     propellantInterlock,
		UmbilicalRetract:  retractor,
		UmbilicalFeedback: umbilicalFeedback,
		UmbilicalActions:  umbilicalActions,
		Interlock:         interlockStatus,
		Holds:             releaseService,
		Range:             rangeService,
		Telemetry:         telemetryService,
		Weather:           weatherService,
		Emergency:         emergencyStates,
		Actuators:         driver,
		Pads:              padRegistry,
		PadConditions:     padConditions,
		Pages:             pages,
	})
	recovery := journal.NewRecoveryCoordinator(flowRecovery, fillRecovery, countdownRecovery)
	if !padConditions.SafeForOperations("LC-1") {
		return nil, fmt.Errorf("launch pad conditions are not operational")
	}
	return &application{handler: server.Handler(), recovery: recovery, flowRecovery: flowRecovery, fillRecovery: fillRecovery, countdownRecovery: countdownRecovery, interlockRecovery: interlockRecovery}, nil
}

func (a *application) recover(ctx context.Context) error {
	if err := a.interlockRecovery.Recover(ctx); err != nil {
		return err
	}
	if err := a.recovery.Recover(ctx); err != nil {
		return err
	}
	if !a.recovery.FlowReadyBeforeControllers() {
		return fmt.Errorf("flow fencing did not precede controller recovery")
	}
	return nil
}

func (a *application) checkpoint() error {
	if err := a.flowRecovery.Save(); err != nil {
		return err
	}
	if err := a.fillRecovery.Save(); err != nil {
		return err
	}
	if err := a.countdownRecovery.Save(); err != nil {
		return err
	}
	return a.interlockRecovery.Save()
}

func initialOperation() model.OperationID {
	return model.NewOperationID()
}
