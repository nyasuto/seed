package server

import (
	"path/filepath"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func TestSaveAndLoadReplay_FileRoundTrip(t *testing.T) {
	sc := loadTutorialScenario(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "replay.json")

	// Create an engine with recording enabled, run a few ticks.
	rng := types.NewCheckpointableRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}
	simulation.EnableRecording(engine)

	for range 5 {
		if _, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}}); err != nil {
			t.Fatalf("Step: %v", err)
		}
	}

	if err := SaveReplay(path, engine); err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	replay, err := LoadReplay(path)
	if err != nil {
		t.Fatalf("LoadReplay: %v", err)
	}

	if replay.Seed != 42 {
		t.Errorf("Seed: got %d, want 42", replay.Seed)
	}
	if replay.ScenarioID != sc.ID {
		t.Errorf("ScenarioID: got %q, want %q", replay.ScenarioID, sc.ID)
	}
	if len(replay.Actions) != 5 {
		t.Errorf("Actions ticks: got %d, want 5", len(replay.Actions))
	}
}

func TestReplay_PlaybackProducesSameResult(t *testing.T) {
	sc := loadTutorialScenario(t)
	dir := t.TempDir()
	replayPath := filepath.Join(dir, "replay.json")

	// Run 1: full game with recording, save replay at end.
	gs1, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	// Use a provider that saves replay when the game ends.
	rp := &replayCapturingProvider{
		replayPath: replayPath,
	}
	originalResult, err := gs1.RunGame(rp)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	// The provider should have saved the replay during OnGameEnd
	// but gs.engine is nil by then. Instead, save via the capturing provider
	// which saves right before game ends.
	// Actually, we need a different approach — save the replay after RunGame
	// using the engine. But engine is nil after RunGame.
	// Let's use a provider that saves during the last OnTickComplete before terminal.
	// Simpler: just run the game, record manually.

	// Re-run with manual recording.
	rng := types.NewCheckpointableRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}
	simulation.EnableRecording(engine)

	maxTicks := 50
	var fullResult simulation.GameResult
	for i := range maxTicks {
		result, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
		if err != nil {
			t.Fatalf("Step at tick %d: %v", i, err)
		}
		if result.Status != simulation.Running {
			fullResult = result
			break
		}
	}

	if err := SaveReplay(replayPath, engine); err != nil {
		t.Fatalf("SaveReplay: %v", err)
	}

	// Run 2: play back the replay.
	gs2, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	replayResult, err := gs2.PlayReplayFrom(replayPath)
	if err != nil {
		t.Fatalf("PlayReplayFrom: %v", err)
	}

	// The replay should produce the same result.
	if fullResult.Status != replayResult.Status {
		t.Errorf("status mismatch: original=%v replay=%v", fullResult.Status, replayResult.Status)
	}
	if fullResult.FinalTick != replayResult.FinalTick {
		t.Errorf("final tick mismatch: original=%d replay=%d", fullResult.FinalTick, replayResult.FinalTick)
	}

	_ = originalResult // used for the initial approach
	_ = rp
}

// replayCapturingProvider is a simple no-op provider (unused in final test but kept for reference).
type replayCapturingProvider struct {
	replayPath string
}

func (p *replayCapturingProvider) ProvideActions(_ scenario.GameSnapshot) ([]simulation.PlayerAction, error) {
	return []simulation.PlayerAction{simulation.NoAction{}}, nil
}
func (p *replayCapturingProvider) OnTickComplete(_ scenario.GameSnapshot) {}
func (p *replayCapturingProvider) OnGameEnd(_ simulation.RunResult)       {}

func TestGameServer_SaveReplayTo_NoEngine(t *testing.T) {
	sc := loadTutorialScenario(t)
	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}
	if err := gs.SaveReplayTo("/tmp/should-not-exist.json"); err == nil {
		t.Fatal("expected error when no engine is active")
	}
}

func TestLoadReplay_BadPath(t *testing.T) {
	_, err := LoadReplay("/nonexistent/path/replay.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadReplay_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := writeFile(path, []byte("not json")); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := LoadReplay(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
