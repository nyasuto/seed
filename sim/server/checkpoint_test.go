package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// checkpointProvider saves a checkpoint at a specified tick, then continues.
type checkpointProvider struct {
	savePath    string
	saveAtTick  int
	engine      **simulation.SimulationEngine
	saved       bool
	tickCount   int
}

func (p *checkpointProvider) ProvideActions(snapshot scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	p.tickCount++
	if !p.saved && int(snapshot.Tick) >= p.saveAtTick {
		if err := SaveCheckpoint(p.savePath, *p.engine); err != nil {
			return nil, err
		}
		p.saved = true
	}
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}

func (p *checkpointProvider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *checkpointProvider) OnGameEnd(_ simulation.RunResult)       {}

func TestSaveAndLoadCheckpoint_FileRoundTrip(t *testing.T) {
	sc := loadTutorialScenario(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "checkpoint.json")

	// Create an engine, run a few ticks, save checkpoint.
	rng := types.NewCheckpointableRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	for range 3 {
		if _, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}}); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}

	if err := SaveCheckpoint(path, engine); err != nil {
		t.Fatalf("SaveCheckpoint: %v", err)
	}

	cp, err := LoadCheckpoint(path)
	if err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}

	if cp.NextBeastID != engine.State.NextBeastID {
		t.Errorf("NextBeastID mismatch: got %d, want %d", cp.NextBeastID, engine.State.NextBeastID)
	}
}

func TestCheckpoint_ResumeProducesSameResult(t *testing.T) {
	sc := loadTutorialScenario(t)
	dir := t.TempDir()
	cpPath := filepath.Join(dir, "checkpoint.json")

	const saveAtTick = 3

	// Run 1: full run (baseline).
	gs1, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	fullResult, err := gs1.RunGame(&lazyAIProvider{sc: sc, seed: 42})
	if err != nil {
		t.Fatalf("RunGame full: %v", err)
	}

	// Run 2: run with checkpoint save at tick N, then load and resume.
	gs2, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	// Use a provider that saves checkpoint at saveAtTick then continues.
	// We need access to gs2.engine, so we use a custom provider that receives
	// the engine pointer indirectly.
	cpProvider := &checkpointProvider{
		savePath:   cpPath,
		saveAtTick: saveAtTick,
		engine:     &gs2.engine,
	}
	// Run the full game with checkpoint saving (it continues after save).
	checkpointResult, err := gs2.RunGame(cpProvider)
	if err != nil {
		t.Fatalf("RunGame with checkpoint: %v", err)
	}

	// Sanity: the full run with checkpoint save should produce same result.
	if fullResult.Result.Status != checkpointResult.Result.Status {
		t.Errorf("full vs checkpoint-save run: status %v vs %v",
			fullResult.Result.Status, checkpointResult.Result.Status)
	}
	if fullResult.TickCount != checkpointResult.TickCount {
		t.Errorf("full vs checkpoint-save run: ticks %d vs %d",
			fullResult.TickCount, checkpointResult.TickCount)
	}

	// Run 3: load checkpoint and resume from tick N.
	gs3, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	if err := gs3.LoadCheckpointFrom(cpPath); err != nil {
		t.Fatalf("LoadCheckpointFrom: %v", err)
	}

	resumeResult, err := gs3.ResumeGame(&lazyAIProvider{sc: sc, seed: 42})
	if err != nil {
		t.Fatalf("ResumeGame: %v", err)
	}

	// The resumed run should produce the same final result as the full run.
	if fullResult.Result.Status != resumeResult.Result.Status {
		t.Errorf("status mismatch: full=%v resume=%v",
			fullResult.Result.Status, resumeResult.Result.Status)
	}
	if fullResult.TickCount != resumeResult.TickCount {
		t.Errorf("tick count mismatch: full=%d resume=%d",
			fullResult.TickCount, resumeResult.TickCount)
	}
	if fullResult.Result.FinalTick != resumeResult.Result.FinalTick {
		t.Errorf("final tick mismatch: full=%d resume=%d",
			fullResult.Result.FinalTick, resumeResult.Result.FinalTick)
	}
	if fullResult.Statistics.PeakChi != resumeResult.Statistics.PeakChi {
		t.Errorf("PeakChi mismatch: full=%f resume=%f",
			fullResult.Statistics.PeakChi, resumeResult.Statistics.PeakChi)
	}
}

func TestGameServer_SaveCheckpointTo_NoEngine(t *testing.T) {
	sc := loadTutorialScenario(t)
	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	if err := gs.SaveCheckpointTo("/tmp/should-not-exist.json"); err == nil {
		t.Fatal("expected error when no engine is active")
	}
}

func TestGameServer_ResumeGame_NoEngine(t *testing.T) {
	sc := loadTutorialScenario(t)
	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	_, err = gs.ResumeGame(&lazyAIProvider{sc: sc, seed: 42})
	if err == nil {
		t.Fatal("expected error when no engine is loaded")
	}
}

func TestLoadCheckpoint_BadPath(t *testing.T) {
	_, err := LoadCheckpoint("/nonexistent/path/checkpoint.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadCheckpoint_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := writeFile(path, []byte("not json")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := LoadCheckpoint(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
