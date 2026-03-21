package render

import (
	"fmt"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/simulation"
	"github.com/nyasuto/seed/core/types"
)

// TestVisual_RenderFullStatus is a manual visual confirmation test.
// Run with: go test -v -run TestVisual ./render/
// It renders the game state at several points during a tutorial game
// and prints the output for human inspection.
func TestVisual_RenderFullStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping visual check in short mode")
	}

	sc := loadTutorialScenario(t)
	rng := types.NewCheckpointableRNG(42)
	engine, err := simulation.NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	// Initial state.
	fmt.Println("=== Initial State (Tick 0) ===")
	fmt.Println(RenderFullStatus(engine.State))

	// Advance a few ticks with NoAction.
	for i := 0; i < 10; i++ {
		_, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
		if err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}

	fmt.Println("=== After 10 Ticks (NoAction) ===")
	fmt.Println(RenderFullStatus(engine.State))

	// Dig a senju_room at (10, 10).
	_, err = engine.Step([]simulation.PlayerAction{simulation.DigRoomAction{
		RoomTypeID: "senju_room",
		Pos:        types.Pos{X: 10, Y: 10},
		Width:      3,
		Height:     3,
	}})
	if err != nil {
		t.Fatalf("DigRoom: %v", err)
	}

	fmt.Println("=== After Digging senju_room at (10,10) ===")
	fmt.Println(RenderFullStatus(engine.State))

	// Summon a Wood beast.
	_, err = engine.Step([]simulation.PlayerAction{simulation.SummonBeastAction{
		Element: types.Wood,
	}})
	if err != nil {
		t.Fatalf("SummonBeast: %v", err)
	}

	fmt.Println("=== After Summoning Wood Beast ===")
	fmt.Println(RenderFullStatus(engine.State))

	// Fast-forward to tick 100 (just before first wave).
	for engine.State.Progress.CurrentTick < 99 {
		res, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
		if err != nil {
			t.Fatalf("Step tick %d: %v", engine.State.Progress.CurrentTick, err)
		}
		if res.Status != simulation.Running {
			fmt.Printf("Game ended early at tick %d: %v\n", res.FinalTick, res.Reason)
			return
		}
	}

	fmt.Println("=== Tick ~100 (Wave Incoming) ===")
	fmt.Println(RenderFullStatus(engine.State))

	// Advance past wave.
	for i := 0; i < 50; i++ {
		res, err := engine.Step([]simulation.PlayerAction{simulation.NoAction{}})
		if err != nil {
			t.Fatalf("Step: %v", err)
		}
		if res.Status != simulation.Running {
			fmt.Printf("=== Game Ended at Tick %d (%v: %s) ===\n", res.FinalTick, res.Status, res.Reason)
			fmt.Println(RenderFullStatus(engine.State))
			return
		}
	}

	fmt.Println("=== Tick ~150 (Post Wave) ===")
	fmt.Println(RenderFullStatus(engine.State))
}

// loadTutorialScenario loads the embedded tutorial scenario.
func loadTutorialScenario(t *testing.T) *scenario.Scenario {
	t.Helper()
	data := []byte(`{
  "id": "tutorial",
  "name": "チュートリアル",
  "description": "基本操作を学ぶための簡単なシナリオ",
  "difficulty": "easy",
  "initial_state": {
    "cave_width": 16, "cave_height": 16,
    "terrain_seed": 1, "terrain_density": 0.05,
    "prebuilt_rooms": [{"type_id": "dragon_hole", "pos": {"x": 7, "y": 7}, "level": 1}],
    "dragon_veins": [{"source_pos": {"x": 7, "y": 7}, "element": "Earth", "flow_rate": 1.0}],
    "starting_chi": 200.0,
    "starting_beasts": [{"species_id": "suiryu", "room_index": 0}]
  },
  "win_conditions": [{"type": "survive_until", "params": {"ticks": 300}}],
  "lose_conditions": [{"type": "core_destroyed"}],
  "wave_schedule": [{"trigger_tick": 100, "difficulty": 0.5, "min_invaders": 1, "max_invaders": 2, "preferred_classes": ["wood_ascetic"], "preferred_goals": []}],
  "events": [],
  "constraints": {"max_rooms": 5, "max_beasts": 3, "max_ticks": 300, "forbidden_room_types": []}
}`)
	sc, err := scenario.LoadScenario(data)
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	return sc
}
