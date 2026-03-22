package controller

import (
	"testing"

	"github.com/nyasuto/seed/core/simulation"
)

func TestCreateCheckpoint_RoundTrip(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController: %v", err)
	}

	// Advance some ticks.
	for i := 0; i < 10; i++ {
		_, err := gc.AdvanceTick()
		if err != nil {
			t.Fatalf("AdvanceTick %d: %v", i, err)
		}
	}

	snapBefore := gc.Snapshot()

	// Create checkpoint.
	cp, err := gc.CreateCheckpoint()
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	// Advance more ticks to change state.
	for i := 0; i < 5; i++ {
		_, _ = gc.AdvanceTick()
	}

	// Restore from checkpoint.
	if err := gc.RestoreFromCheckpoint(cp); err != nil {
		t.Fatalf("RestoreFromCheckpoint: %v", err)
	}

	snapAfter := gc.Snapshot()
	if snapAfter.Tick != snapBefore.Tick {
		t.Errorf("tick = %d, want %d", snapAfter.Tick, snapBefore.Tick)
	}
	if gc.State() != Playing {
		t.Errorf("state = %d, want Playing", gc.State())
	}
}

func TestNewGameControllerFromCheckpoint(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController: %v", err)
	}

	for i := 0; i < 10; i++ {
		_, _ = gc.AdvanceTick()
	}

	snapBefore := gc.Snapshot()

	cp, err := gc.CreateCheckpoint()
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	gc2, err := NewGameControllerFromCheckpoint(cp, tutorialJSON)
	if err != nil {
		t.Fatalf("NewGameControllerFromCheckpoint: %v", err)
	}

	snapAfter := gc2.Snapshot()
	if snapAfter.Tick != snapBefore.Tick {
		t.Errorf("tick = %d, want %d", snapAfter.Tick, snapBefore.Tick)
	}
	if snapAfter.CoreHP != snapBefore.CoreHP {
		t.Errorf("coreHP = %d, want %d", snapAfter.CoreHP, snapBefore.CoreHP)
	}
}

func TestScenarioJSON_Preserved(t *testing.T) {
	gc, err := NewGameController(tutorialJSON, 42)
	if err != nil {
		t.Fatalf("NewGameController: %v", err)
	}

	json := gc.ScenarioJSON()
	if len(json) == 0 {
		t.Fatal("ScenarioJSON() returned empty")
	}
	if string(json) != string(tutorialJSON) {
		t.Error("ScenarioJSON does not match original")
	}
}

func TestCheckpoint_ContinueDeterminism(t *testing.T) {
	// Path A: run 10 ticks, checkpoint, run 5 more.
	gc1, _ := NewGameController(tutorialJSON, 42)
	for i := 0; i < 10; i++ {
		_, _ = gc1.AdvanceTick()
	}
	cp, err := gc1.CreateCheckpoint()
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
	for i := 0; i < 5; i++ {
		_, _ = gc1.AdvanceTick()
	}
	snapA := gc1.Snapshot()

	// Path B: restore from checkpoint, run 5 ticks.
	gc2, err := NewGameControllerFromCheckpoint(cp, tutorialJSON)
	if err != nil {
		t.Fatalf("NewGameControllerFromCheckpoint: %v", err)
	}
	for i := 0; i < 5; i++ {
		result, err := gc2.AdvanceTick()
		if err != nil {
			t.Fatalf("AdvanceTick %d: %v", i, err)
		}
		if result.Status != simulation.Running {
			break
		}
	}
	snapB := gc2.Snapshot()

	if snapA.Tick != snapB.Tick {
		t.Errorf("tick: A=%d B=%d", snapA.Tick, snapB.Tick)
	}
	if snapA.CoreHP != snapB.CoreHP {
		t.Errorf("coreHP: A=%d B=%d", snapA.CoreHP, snapB.CoreHP)
	}
	if snapA.ChiPoolBalance != snapB.ChiPoolBalance {
		t.Errorf("chiPool: A=%f B=%f", snapA.ChiPoolBalance, snapB.ChiPoolBalance)
	}
}
