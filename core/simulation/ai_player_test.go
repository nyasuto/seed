package simulation

import (
	"encoding/json"
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
)

// tutorialScenario returns a simple scenario that SimpleAI should be able to
// clear by surviving 10 ticks with no invasion waves.
func tutorialScenario() *scenario.Scenario {
	return &scenario.Scenario{
		ID:         "tutorial",
		Name:       "Tutorial",
		Difficulty: "easy",
		InitialState: scenario.InitialState{
			CaveWidth:      20,
			CaveHeight:     20,
			TerrainSeed:    42,
			TerrainDensity: 0.0,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
			},
			DragonVeins: []scenario.DragonVeinPlacement{
				{SourcePos: types.Pos{X: 5, Y: 7}, Element: types.Earth, FlowRate: 5.0},
			},
			StartingChi: 200.0,
		},
		WinConditions: []scenario.ConditionDef{
			{Type: "survive_until", Params: json.RawMessage(`{"ticks": 10}`)},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed"},
		},
	}
}

func TestSimpleAIPlayer_DecideActions_NoAction(t *testing.T) {
	sc := tutorialScenario()
	// Set very low starting chi so AI cannot afford anything.
	sc.InitialState.StartingChi = 0.1
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)
	snapshot := BuildSnapshot(engine.State)
	actions := ai.DecideActions(snapshot)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].ActionType() != "no_action" {
		t.Errorf("expected no_action, got %s", actions[0].ActionType())
	}
}

func TestSimpleAIPlayer_TutorialScenarioClear(t *testing.T) {
	sc := tutorialScenario()
	rng := types.NewSeededRNG(42)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)
	result, err := engine.Run(50, func(snap scenario.GameSnapshot) []PlayerAction {
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != Won {
		t.Errorf("expected Won, got %v (reason: %s, tick: %d)", result.Status, result.Reason, result.FinalTick)
	}
}

func TestRandomAIPlayer_NoCrash(t *testing.T) {
	sc := tutorialScenario()
	engineRNG := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, engineRNG)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	aiRNG := types.NewSeededRNG(99)
	ai := NewRandomAIPlayer(engine.State, aiRNG)

	// Run for 30 ticks; the test passes if no panic or error occurs.
	result, err := engine.Run(30, func(snap scenario.GameSnapshot) []PlayerAction {
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The game should either be won or lost or still running — just not error.
	t.Logf("result: status=%v tick=%d reason=%s", result.Status, result.FinalTick, result.Reason)
}

func TestRandomAIPlayer_Deterministic(t *testing.T) {
	run := func(seed int64) GameResult {
		sc := tutorialScenario()
		engineRNG := types.NewSeededRNG(seed)
		engine, err := NewSimulationEngine(sc, engineRNG)
		if err != nil {
			t.Fatalf("NewSimulationEngine: %v", err)
		}
		aiRNG := types.NewSeededRNG(seed + 1000)
		ai := NewRandomAIPlayer(engine.State, aiRNG)
		result, err := engine.Run(20, func(snap scenario.GameSnapshot) []PlayerAction {
			return ai.DecideActions(snap)
		})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		return result
	}

	r1 := run(42)
	r2 := run(42)

	if r1.Status != r2.Status || r1.FinalTick != r2.FinalTick || r1.Reason != r2.Reason {
		t.Errorf("non-deterministic: run1=%+v, run2=%+v", r1, r2)
	}
}

func TestSimpleAIPlayer_Interface(t *testing.T) {
	// Verify that SimpleAIPlayer and RandomAIPlayer satisfy the AIPlayer interface.
	sc := tutorialScenario()
	rng := types.NewSeededRNG(1)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	var _ AIPlayer = NewSimpleAIPlayer(engine.State)
	var _ AIPlayer = NewRandomAIPlayer(engine.State, types.NewSeededRNG(2))
}

func TestSimpleAIPlayer_RespectsMaxRooms(t *testing.T) {
	sc := tutorialScenario()
	// Allow building rooms by giving enough chi.
	sc.InitialState.StartingChi = 500.0
	// Set MaxRooms to 1. The prebuilt dragon_hole counts as 1 room.
	sc.Constraints.MaxRooms = 1

	rng := types.NewSeededRNG(1)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)
	snapshot := BuildSnapshot(engine.State)
	actions := ai.DecideActions(snapshot)

	// AI should not attempt to build a room since MaxRooms is reached.
	for _, a := range actions {
		if a.ActionType() == "dig_room" {
			t.Error("SimpleAIPlayer should not attempt dig_room when MaxRooms is reached")
		}
	}
}

func TestValidateAction_DigRoom_MaxRoomsAfterBuilding(t *testing.T) {
	sc := tutorialScenario()
	sc.InitialState.StartingChi = 500.0
	// Allow up to 5 rooms total.
	sc.Constraints.MaxRooms = 5

	rng := types.NewSeededRNG(1)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}
	state := engine.State

	// Build rooms until MaxRooms is reached.
	positions := []types.Pos{
		{X: 1, Y: 1},
		{X: 5, Y: 1},
		{X: 9, Y: 1},
		{X: 13, Y: 1},
	}
	for _, pos := range positions {
		action := DigRoomAction{
			RoomTypeID: "trap_room",
			Pos:        pos,
			Width:      3,
			Height:     3,
		}
		if _, err := ApplyAction(action, state); err != nil {
			t.Fatalf("ApplyAction DigRoom at %v: %v", pos, err)
		}
	}

	// Now we should have 5 rooms (1 prebuilt + 4 built).
	if got := len(state.Cave.Rooms); got != 5 {
		t.Fatalf("room count = %d, want 5", got)
	}

	// 6th room should be rejected.
	action := DigRoomAction{
		RoomTypeID: "trap_room",
		Pos:        types.Pos{X: 1, Y: 10},
		Width:      3,
		Height:     3,
	}
	err = ValidateAction(action, state)
	if err == nil {
		t.Fatal("expected error when MaxRooms reached")
	}
	if got := err.Error(); got != "max rooms reached: 5/5" {
		t.Errorf("error = %q, want %q", got, "max rooms reached: 5/5")
	}
}
