package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

func newTestEngine(t *testing.T) (*simulation.SimulationEngine, []byte) {
	t.Helper()
	scenarioJSON := loadTestScenario(t)
	sc := loadScenario(t, scenarioJSON)
	rng := types.NewCheckpointableRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}
	return engine, scenarioJSON
}

func TestSaveCheckpoint_RoundTrip(t *testing.T) {
	engine, scenarioJSON := newTestEngine(t)

	// Advance a few ticks.
	for i := 0; i < 5; i++ {
		_, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
		if err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}

	tickBefore := simulation.BuildSnapshot(engine.State).Tick

	// Save.
	dir := t.TempDir()
	path := filepath.Join(dir, "test_save.json")
	err := SaveCheckpoint(path, engine, scenarioJSON, "tutorial")
	if err != nil {
		t.Fatalf("SaveCheckpoint: %v", err)
	}

	// Verify file is valid JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("save file is not valid JSON")
	}

	// Load.
	sf, err := LoadSaveFile(path)
	if err != nil {
		t.Fatalf("LoadSaveFile: %v", err)
	}

	if sf.ScenarioID != "tutorial" {
		t.Errorf("scenario ID = %q, want %q", sf.ScenarioID, "tutorial")
	}
	if sf.SavedAt.IsZero() {
		t.Error("saved_at is zero")
	}
	if sf.Checkpoint == nil {
		t.Fatal("checkpoint is nil")
	}

	// Restore engine and verify tick matches.
	sc := loadScenario(t, sf.ScenarioJSON)
	restored, err := simulation.RestoreCheckpoint(sf.Checkpoint, sc)
	if err != nil {
		t.Fatalf("RestoreCheckpoint: %v", err)
	}
	tickAfter := simulation.BuildSnapshot(restored.State).Tick
	if tickAfter != tickBefore {
		t.Errorf("tick after restore = %d, want %d", tickAfter, tickBefore)
	}
}

func TestSaveCheckpoint_SnapshotMatch(t *testing.T) {
	engine, scenarioJSON := newTestEngine(t)

	// Advance some ticks.
	for i := 0; i < 10; i++ {
		_, _ = engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
	}

	snapBefore := simulation.BuildSnapshot(engine.State)

	// Save and restore.
	dir := t.TempDir()
	path := filepath.Join(dir, "snap_test.json")
	if err := SaveCheckpoint(path, engine, scenarioJSON, "tutorial"); err != nil {
		t.Fatalf("SaveCheckpoint: %v", err)
	}

	sf, err := LoadSaveFile(path)
	if err != nil {
		t.Fatalf("LoadSaveFile: %v", err)
	}

	sc := loadScenario(t, sf.ScenarioJSON)
	restored, err := simulation.RestoreCheckpoint(sf.Checkpoint, sc)
	if err != nil {
		t.Fatalf("RestoreCheckpoint: %v", err)
	}

	snapAfter := simulation.BuildSnapshot(restored.State)

	if snapAfter.Tick != snapBefore.Tick {
		t.Errorf("tick: got %d, want %d", snapAfter.Tick, snapBefore.Tick)
	}
	if snapAfter.CoreHP != snapBefore.CoreHP {
		t.Errorf("coreHP: got %d, want %d", snapAfter.CoreHP, snapBefore.CoreHP)
	}
	if snapAfter.ChiPoolBalance != snapBefore.ChiPoolBalance {
		t.Errorf("chiPool: got %f, want %f", snapAfter.ChiPoolBalance, snapBefore.ChiPoolBalance)
	}
}

func TestListSaves_NonexistentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")
	saves, err := ListSaves(dir)
	if err != nil {
		t.Fatalf("ListSaves: %v", err)
	}
	if len(saves) != 0 {
		t.Errorf("expected empty list, got %d entries", len(saves))
	}
}

func TestListSaves_SortedNewestFirst(t *testing.T) {
	engine, scenarioJSON := newTestEngine(t)
	dir := t.TempDir()

	// Create two saves with different timestamps.
	path1 := filepath.Join(dir, "save_old.json")
	path2 := filepath.Join(dir, "save_new.json")

	sf1 := SaveFile{
		ScenarioJSON: scenarioJSON,
		SavedAt:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ScenarioID:   "tutorial",
	}
	cp, err := simulation.CreateCheckpoint(engine)
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}
	sf1.Checkpoint = cp
	sf2 := SaveFile{
		ScenarioJSON: scenarioJSON,
		Checkpoint:   cp,
		SavedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		ScenarioID:   "tutorial",
	}

	writeJSON(t, path1, sf1)
	writeJSON(t, path2, sf2)

	saves, err := ListSaves(dir)
	if err != nil {
		t.Fatalf("ListSaves: %v", err)
	}
	if len(saves) != 2 {
		t.Fatalf("expected 2 saves, got %d", len(saves))
	}
	if saves[0].Filename != "save_new.json" {
		t.Errorf("first entry = %q, want save_new.json", saves[0].Filename)
	}
	if saves[1].Filename != "save_old.json" {
		t.Errorf("second entry = %q, want save_old.json", saves[1].Filename)
	}
}

func TestGenerateFilename(t *testing.T) {
	name := GenerateFilename()
	if len(name) == 0 {
		t.Fatal("empty filename")
	}
	if filepath.Ext(name) != ".json" {
		t.Errorf("extension = %q, want .json", filepath.Ext(name))
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
