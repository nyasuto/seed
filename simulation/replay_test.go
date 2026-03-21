package simulation

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
)

func TestMarshalReplay_RoundTrip(t *testing.T) {
	original := &Replay{
		Seed:       42,
		ScenarioID: "test_scenario",
		Actions: map[types.Tick][]PlayerAction{
			0: {NoAction{}},
			1: {DigRoomAction{RoomTypeID: "fire_room", Pos: types.Pos{X: 3, Y: 4}, Width: 3, Height: 3}},
			3: {
				SummonBeastAction{Element: types.Wood},
				PlaceBeastAction{SpeciesID: "kodama", RoomID: 1},
			},
			5: {UpgradeRoomAction{RoomID: 1}},
			7: {DigCorridorAction{FromRoomID: 1, ToRoomID: 2}},
			9: {EvolveBeastAction{BeastID: 1}},
		},
	}

	data, err := MarshalReplay(original)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	restored, err := UnmarshalReplay(data)
	if err != nil {
		t.Fatalf("UnmarshalReplay: %v", err)
	}

	if restored.Seed != original.Seed {
		t.Errorf("Seed = %d, want %d", restored.Seed, original.Seed)
	}
	if restored.ScenarioID != original.ScenarioID {
		t.Errorf("ScenarioID = %q, want %q", restored.ScenarioID, original.ScenarioID)
	}
	if len(restored.Actions) != len(original.Actions) {
		t.Fatalf("Actions length = %d, want %d", len(restored.Actions), len(original.Actions))
	}

	for tick, origActions := range original.Actions {
		restoredActions, ok := restored.Actions[tick]
		if !ok {
			t.Errorf("missing actions for tick %d", tick)
			continue
		}
		if len(restoredActions) != len(origActions) {
			t.Errorf("tick %d: action count = %d, want %d", tick, len(restoredActions), len(origActions))
			continue
		}
		for i, orig := range origActions {
			if !reflect.DeepEqual(orig, restoredActions[i]) {
				t.Errorf("tick %d action %d: got %+v, want %+v", tick, i, restoredActions[i], orig)
			}
		}
	}
}

func TestMarshalReplay_ValidJSON(t *testing.T) {
	r := &Replay{
		Seed:       1,
		ScenarioID: "s1",
		Actions: map[types.Tick][]PlayerAction{
			0: {NoAction{}},
		},
	}

	data, err := MarshalReplay(r)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	// Verify it's valid JSON by parsing into a generic map.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if m["seed"] != float64(1) {
		t.Errorf("seed = %v, want 1", m["seed"])
	}
	if m["scenario_id"] != "s1" {
		t.Errorf("scenario_id = %v, want s1", m["scenario_id"])
	}
}

func TestUnmarshalReplay_InvalidJSON(t *testing.T) {
	_, err := UnmarshalReplay([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestUnmarshalReplay_UnknownActionType(t *testing.T) {
	data := []byte(`{"seed":1,"scenario_id":"s","actions":{"0":[{"type":"unknown_type","data":{}}]}}`)
	_, err := UnmarshalReplay(data)
	if err == nil {
		t.Fatal("expected error for unknown action type")
	}
}

func TestMarshalReplay_EmptyActions(t *testing.T) {
	r := &Replay{
		Seed:       99,
		ScenarioID: "empty",
		Actions:    map[types.Tick][]PlayerAction{},
	}

	data, err := MarshalReplay(r)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	restored, err := UnmarshalReplay(data)
	if err != nil {
		t.Fatalf("UnmarshalReplay: %v", err)
	}

	if len(restored.Actions) != 0 {
		t.Errorf("expected empty actions, got %d", len(restored.Actions))
	}
}

func TestRecordReplay_PlayReplay_SameResult(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewCheckpointableRNG(42)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	EnableRecording(engine)

	actions := []PlayerAction{NoAction{}}
	result1, err := engine.Run(5, func(_ scenario.GameSnapshot) []PlayerAction {
		return actions
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	replay, err := RecordReplay(engine)
	if err != nil {
		t.Fatalf("RecordReplay: %v", err)
	}

	// Marshal and unmarshal to simulate save/load.
	data, err := MarshalReplay(replay)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	restored, err := UnmarshalReplay(data)
	if err != nil {
		t.Fatalf("UnmarshalReplay: %v", err)
	}

	result2, err := PlayReplay(restored, sc)
	if err != nil {
		t.Fatalf("PlayReplay: %v", err)
	}

	assertGameResultEqual(t, result1, result2)
}

func TestRecordReplay_PlayReplay_WithWinCondition(t *testing.T) {
	sc := minimalScenario()
	sc.WinConditions = []scenario.ConditionDef{
		{Type: "survive_until", Params: map[string]any{"ticks": float64(5)}},
	}
	rng := types.NewCheckpointableRNG(99)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	EnableRecording(engine)

	result1, err := engine.Run(100, func(_ scenario.GameSnapshot) []PlayerAction {
		return []PlayerAction{NoAction{}}
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result1.Status != Won {
		t.Fatalf("expected Won status, got %v", result1.Status)
	}

	replay, err := RecordReplay(engine)
	if err != nil {
		t.Fatalf("RecordReplay: %v", err)
	}

	result2, err := PlayReplay(replay, sc)
	if err != nil {
		t.Fatalf("PlayReplay: %v", err)
	}

	assertGameResultEqual(t, result1, result2)
}

func TestRecordReplay_PlayReplay_MultipleRuns(t *testing.T) {
	// Replaying the same replay multiple times should produce identical results.
	sc := minimalScenario()
	rng := types.NewCheckpointableRNG(7)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	EnableRecording(engine)

	result1, err := engine.Run(10, func(_ scenario.GameSnapshot) []PlayerAction {
		return []PlayerAction{NoAction{}}
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	replay, err := RecordReplay(engine)
	if err != nil {
		t.Fatalf("RecordReplay: %v", err)
	}

	for i := range 3 {
		result, err := PlayReplay(replay, sc)
		if err != nil {
			t.Fatalf("PlayReplay iteration %d: %v", i, err)
		}
		assertGameResultEqual(t, result1, result)
	}
}

func TestMarshalReplay_RoundTrip_ThenPlay(t *testing.T) {
	// Full round trip: record → marshal → unmarshal → play → same result.
	sc := minimalScenario()
	rng := types.NewCheckpointableRNG(123)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	EnableRecording(engine)

	result1, err := engine.Run(8, func(_ scenario.GameSnapshot) []PlayerAction {
		return []PlayerAction{NoAction{}}
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	replay, err := RecordReplay(engine)
	if err != nil {
		t.Fatalf("RecordReplay: %v", err)
	}

	data, err := MarshalReplay(replay)
	if err != nil {
		t.Fatalf("MarshalReplay: %v", err)
	}

	restored, err := UnmarshalReplay(data)
	if err != nil {
		t.Fatalf("UnmarshalReplay: %v", err)
	}

	// Verify replay metadata survived the round trip.
	if restored.Seed != replay.Seed {
		t.Errorf("Seed = %d, want %d", restored.Seed, replay.Seed)
	}
	if restored.ScenarioID != replay.ScenarioID {
		t.Errorf("ScenarioID = %q, want %q", restored.ScenarioID, replay.ScenarioID)
	}
	if len(restored.Actions) != len(replay.Actions) {
		t.Errorf("Actions count = %d, want %d", len(restored.Actions), len(replay.Actions))
	}

	result2, err := PlayReplay(restored, sc)
	if err != nil {
		t.Fatalf("PlayReplay: %v", err)
	}

	assertGameResultEqual(t, result1, result2)
}

func TestRecordReplay_NonCheckpointableRNG(t *testing.T) {
	sc := minimalScenario()
	// NewSeededRNG does not implement CheckpointableRNG.
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	_, err = RecordReplay(engine)
	if err != ErrRNGNotCheckpointable {
		t.Errorf("expected ErrRNGNotCheckpointable, got %v", err)
	}
}

// assertGameResultEqual is a test helper that checks all fields of GameResult.
func assertGameResultEqual(t *testing.T, want, got GameResult) {
	t.Helper()
	if got.Status != want.Status {
		t.Errorf("Status: got %v, want %v", got.Status, want.Status)
	}
	if got.FinalTick != want.FinalTick {
		t.Errorf("FinalTick: got %d, want %d", got.FinalTick, want.FinalTick)
	}
	if got.Reason != want.Reason {
		t.Errorf("Reason: got %q, want %q", got.Reason, want.Reason)
	}
}
