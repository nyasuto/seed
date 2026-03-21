package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nyasuto/seed/core/simulation"
)

func TestLoadBuiltinScenario_Tutorial(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario(tutorial): %v", err)
	}
	if sc.ID != "tutorial" {
		t.Errorf("expected ID=tutorial, got %s", sc.ID)
	}
	if sc.Difficulty != "easy" {
		t.Errorf("expected Difficulty=easy, got %s", sc.Difficulty)
	}
}

func TestLoadBuiltinScenario_Standard(t *testing.T) {
	sc, err := LoadBuiltinScenario("standard")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario(standard): %v", err)
	}
	if sc.ID != "standard" {
		t.Errorf("expected ID=standard, got %s", sc.ID)
	}
	if sc.Difficulty != "normal" {
		t.Errorf("expected Difficulty=normal, got %s", sc.Difficulty)
	}
}

func TestLoadBuiltinScenario_Unknown(t *testing.T) {
	_, err := LoadBuiltinScenario("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown builtin scenario")
	}
}

func TestLoadScenarioFromFile(t *testing.T) {
	// Write a minimal scenario JSON to a temp file.
	dir := t.TempDir()
	path := filepath.Join(dir, "test_scenario.json")

	data := []byte(`{
		"id": "file_test",
		"name": "File Test",
		"difficulty": "easy",
		"initial_state": {
			"cave_width": 16, "cave_height": 16,
			"terrain_seed": 1, "terrain_density": 0.0,
			"prebuilt_rooms": [{"type_id":"dragon_hole","pos":{"x":5,"y":5},"level":1}],
			"dragon_veins": [{"source_pos":{"x":5,"y":7},"element":"Earth","flow_rate":1.0}],
			"starting_chi": 100.0, "starting_beasts": []
		},
		"win_conditions": [{"type":"survive_until","params":{"ticks":10}}],
		"lose_conditions": [{"type":"core_destroyed"}],
		"wave_schedule": [], "events": [],
		"constraints": {"max_rooms":5,"max_beasts":3,"max_ticks":50}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sc, err := LoadScenarioFromFile(path)
	if err != nil {
		t.Fatalf("LoadScenarioFromFile: %v", err)
	}
	if sc.ID != "file_test" {
		t.Errorf("expected ID=file_test, got %s", sc.ID)
	}
}

func TestLoadScenarioFromFile_NotFound(t *testing.T) {
	_, err := LoadScenarioFromFile("/nonexistent/path/scenario.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadScenario_BuiltinName(t *testing.T) {
	sc, err := LoadScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadScenario(tutorial): %v", err)
	}
	if sc.ID != "tutorial" {
		t.Errorf("expected ID=tutorial, got %s", sc.ID)
	}
}

func TestLoadScenario_FilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.json")

	data := []byte(`{
		"id": "custom",
		"name": "Custom",
		"difficulty": "easy",
		"initial_state": {
			"cave_width": 16, "cave_height": 16,
			"terrain_seed": 1, "terrain_density": 0.0,
			"prebuilt_rooms": [{"type_id":"dragon_hole","pos":{"x":5,"y":5},"level":1}],
			"dragon_veins": [{"source_pos":{"x":5,"y":7},"element":"Earth","flow_rate":1.0}],
			"starting_chi": 100.0, "starting_beasts": []
		},
		"win_conditions": [{"type":"survive_until","params":{"ticks":10}}],
		"lose_conditions": [{"type":"core_destroyed"}],
		"wave_schedule": [], "events": [],
		"constraints": {"max_rooms":5,"max_beasts":3,"max_ticks":50}
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	sc, err := LoadScenario(path)
	if err != nil {
		t.Fatalf("LoadScenario(path): %v", err)
	}
	if sc.ID != "custom" {
		t.Errorf("expected ID=custom, got %s", sc.ID)
	}
}

func TestBuiltinScenarioNames(t *testing.T) {
	names := BuiltinScenarioNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 builtin names, got %d", len(names))
	}
	has := map[string]bool{}
	for _, n := range names {
		has[n] = true
	}
	if !has["tutorial"] {
		t.Error("missing tutorial in builtin names")
	}
	if !has["standard"] {
		t.Error("missing standard in builtin names")
	}
}

// TestLoadBuiltinScenario_TutorialRunGame verifies that a built-in tutorial
// scenario can be loaded and used to start a game via GameServer.
func TestLoadBuiltinScenario_TutorialRunGame(t *testing.T) {
	sc, err := LoadBuiltinScenario("tutorial")
	if err != nil {
		t.Fatalf("LoadBuiltinScenario: %v", err)
	}

	gs, err := NewGameServer(sc, 42)
	if err != nil {
		t.Fatalf("NewGameServer: %v", err)
	}

	provider := &lazyAIProvider{sc: sc, seed: 42}
	result, err := gs.RunGame(provider)
	if err != nil {
		t.Fatalf("RunGame: %v", err)
	}

	if result.Result.Status != simulation.Won && result.Result.Status != simulation.Lost {
		t.Errorf("expected terminal status, got %v", result.Result.Status)
	}
	if result.TickCount == 0 {
		t.Error("expected TickCount > 0")
	}
}
