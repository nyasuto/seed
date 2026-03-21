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
