package simulation

import (
	"strings"
	"testing"

	"github.com/ponpoko/chaosseed-core/economy"
	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/invasion"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/senju"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// buildMinimalEngine creates a minimal SimulationEngine for ASCII rendering tests.
func buildMinimalEngine(t *testing.T) *SimulationEngine {
	t.Helper()
	cave, err := world.NewCave(8, 8)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	roomReg, err := world.LoadDefaultRoomTypes()
	if err != nil {
		t.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	// Add a room so we have room cells to render.
	_, err = cave.AddRoom("wood_garden", types.Pos{X: 2, Y: 2}, 3, 3, []world.RoomEntrance{
		{Pos: types.Pos{X: 3, Y: 4}, Dir: types.South},
	})
	if err != nil {
		t.Fatalf("AddRoom: %v", err)
	}

	veins := []*fengshui.DragonVein{}
	chiEngine := fengshui.NewChiFlowEngine(cave, veins, roomReg, fengshui.DefaultFlowParams())

	chiPool := economy.NewChiPool(100)
	_ = chiPool.Deposit(50, economy.Supply, "init", 0)
	econEngine := economy.NewEconomyEngine(
		chiPool,
		economy.DefaultSupplyParams(),
		economy.DefaultCostParams(),
		economy.DefaultDeficitParams(),
		economy.DefaultConstructionCost(),
		economy.DefaultBeastCost(),
	)

	progress := &scenario.ScenarioProgress{
		ScenarioID:  "test_scenario",
		CurrentTick: 10,
		CoreHP:      100,
	}

	state := &GameState{
		Cave:             cave,
		RoomTypeRegistry: roomReg,
		ChiFlowEngine:    chiEngine,
		Beasts:           nil,
		EconomyEngine:    econEngine,
		Scenario:         &scenario.Scenario{ID: "test_scenario"},
		Progress:         progress,
		ScoreParams:      fengshui.DefaultScoreParams(),
	}

	return &SimulationEngine{
		State:    state,
		Executor: NewCommandExecutor(),
	}
}

func TestRenderFullStatus_BasicOutput(t *testing.T) {
	engine := buildMinimalEngine(t)
	output := RenderFullStatus(engine)

	if output == "" {
		t.Fatal("RenderFullStatus returned empty string")
	}

	// Should contain status panel header.
	if !strings.Contains(output, "--- Status ---") {
		t.Error("missing status panel header")
	}

	// Should contain tick info.
	if !strings.Contains(output, "Tick: 10") {
		t.Error("missing tick display")
	}

	// Should contain core HP.
	if !strings.Contains(output, "Core HP: 100") {
		t.Error("missing core HP display")
	}

	// Should contain chi pool balance.
	if !strings.Contains(output, "Chi Pool: 50.0") {
		t.Error("missing chi pool balance")
	}

	// Should contain scenario ID.
	if !strings.Contains(output, "test_scenario") {
		t.Error("missing scenario ID")
	}
}

func TestRenderFullStatus_WithBeasts(t *testing.T) {
	engine := buildMinimalEngine(t)
	engine.State.Beasts = []*senju.Beast{
		{ID: 1, Element: types.Wood, RoomID: 1, HP: 10, State: senju.Idle},
		{ID: 2, Element: types.Fire, RoomID: 1, HP: 5, State: senju.Stunned},
	}

	output := RenderFullStatus(engine)

	// Map should show beast tile for room 1 (not stunned beast excluded from alive).
	// Status should report beasts.
	if !strings.Contains(output, "Beasts: 2 (alive: 1, stunned: 1)") {
		t.Error("missing or incorrect beast count in status")
	}
}

func TestRenderFullStatus_WithInvaders(t *testing.T) {
	engine := buildMinimalEngine(t)
	engine.State.Waves = []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 1, State: invasion.Advancing, CurrentRoomID: 1},
				{ID: 2, State: invasion.Defeated, CurrentRoomID: 0},
			},
		},
		{
			ID:    2,
			State: invasion.Completed,
			Invaders: []*invasion.Invader{
				{ID: 3, State: invasion.Defeated, CurrentRoomID: 0},
			},
		},
	}

	output := RenderFullStatus(engine)

	if !strings.Contains(output, "Waves: 2 total, 1 active, 1 completed") {
		t.Errorf("missing or incorrect wave count, got:\n%s", output)
	}
	if !strings.Contains(output, "Invaders: 3 total, 1 active") {
		t.Errorf("missing or incorrect invader count, got:\n%s", output)
	}

	// Room 1 should show invader tile ">>" (priority over beasts and chi).
	lines := strings.Split(output, "\n")
	mapHasInvaderTile := false
	for _, line := range lines {
		if strings.Contains(line, ">>") {
			mapHasInvaderTile = true
			break
		}
	}
	if !mapHasInvaderTile {
		t.Error("map should contain invader tile '>>'")
	}
}

func TestRenderFullStatus_InvaderPriorityOverBeasts(t *testing.T) {
	engine := buildMinimalEngine(t)

	// Both beasts and invaders in room 1.
	engine.State.Beasts = []*senju.Beast{
		{ID: 1, Element: types.Wood, RoomID: 1, HP: 10, State: senju.Idle},
	}
	engine.State.Waves = []*invasion.InvasionWave{
		{
			ID:    1,
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{ID: 1, State: invasion.Fighting, CurrentRoomID: 1},
			},
		},
	}

	output := RenderFullStatus(engine)
	lines := strings.Split(output, "\n")

	// Room cells should show invader tile "XX" not beast tile "WW".
	hasXX := false
	hasWW := false
	for _, line := range lines {
		if strings.Contains(line, "--- Status ---") {
			break // Stop at status panel
		}
		if strings.Contains(line, "XX") {
			hasXX = true
		}
		if strings.Contains(line, "WW") {
			hasWW = true
		}
	}
	if !hasXX {
		t.Error("room with invaders should show 'XX' (fighting)")
	}
	if hasWW {
		t.Error("room with invaders should NOT show beast tile 'WW'")
	}
}

func TestRenderFullStatus_DamageAndEvolutions(t *testing.T) {
	engine := buildMinimalEngine(t)
	engine.State.TotalDamageDealt = 42
	engine.State.TotalDamageReceived = 15
	engine.State.EvolutionCount = 3

	output := RenderFullStatus(engine)

	if !strings.Contains(output, "Damage: dealt 42, received 15") {
		t.Error("missing or incorrect damage stats")
	}
	if !strings.Contains(output, "Evolutions: 3") {
		t.Error("missing or incorrect evolution count")
	}
}

func TestRenderFullStatus_DeficitTracking(t *testing.T) {
	engine := buildMinimalEngine(t)
	engine.State.TotalDeficitTicks = 5
	engine.State.ConsecutiveDeficitTicks = 2

	output := RenderFullStatus(engine)

	if !strings.Contains(output, "Deficit: 5 total ticks, 2 consecutive") {
		t.Error("missing or incorrect deficit display")
	}
}

func TestRenderFullStatus_NilSafety(t *testing.T) {
	// Minimal state with nil subsystems should not panic.
	state := &GameState{
		Progress: &scenario.ScenarioProgress{CoreHP: 50},
		Scenario: &scenario.Scenario{ID: "nil_test"},
	}
	engine := &SimulationEngine{
		State:    state,
		Executor: NewCommandExecutor(),
	}

	output := RenderFullStatus(engine)
	if !strings.Contains(output, "Core HP: 50") {
		t.Error("nil-safe render should still show core HP")
	}
}

func TestRenderFullStatus_MultipleBeastsInRoom(t *testing.T) {
	engine := buildMinimalEngine(t)
	engine.State.Beasts = []*senju.Beast{
		{ID: 1, Element: types.Fire, RoomID: 1, HP: 10, State: senju.Idle},
		{ID: 2, Element: types.Fire, RoomID: 1, HP: 8, State: senju.Idle},
		{ID: 3, Element: types.Fire, RoomID: 1, HP: 6, State: senju.Idle},
	}

	output := RenderFullStatus(engine)
	lines := strings.Split(output, "\n")

	// Room cells should show "3F" for 3 fire beasts.
	found := false
	for _, line := range lines {
		if strings.Contains(line, "--- Status ---") {
			break
		}
		if strings.Contains(line, "3F") {
			found = true
			break
		}
	}
	if !found {
		t.Error("3 fire beasts in room should show '3F' tile")
	}
}

func TestRenderFullStatus_WithChiOverlay(t *testing.T) {
	engine := buildMinimalEngine(t)

	// Set chi in room 1 to 50% fill.
	engine.State.ChiFlowEngine.RoomChi[1] = &fengshui.RoomChi{
		Current:  50,
		Capacity: 100,
	}

	output := RenderFullStatus(engine)
	lines := strings.Split(output, "\n")

	// Room cells should show chi overlay ▒▒ (34-66%).
	found := false
	for _, line := range lines {
		if strings.Contains(line, "--- Status ---") {
			break
		}
		if strings.Contains(line, "▒▒") {
			found = true
			break
		}
	}
	if !found {
		t.Error("room with 50% chi should show '▒▒' tile")
	}
}

func TestRenderFullStatus_WithScenarioEngine(t *testing.T) {
	// Use the full NewSimulationEngine to verify integration.
	sc := minimalScenario()
	rng := types.NewSeededRNG(42)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	output := RenderFullStatus(engine)
	if output == "" {
		t.Fatal("RenderFullStatus on full engine returned empty")
	}
	if !strings.Contains(output, "--- Status ---") {
		t.Error("missing status panel")
	}
	if !strings.Contains(output, "Tick: 0") {
		t.Error("initial tick should be 0")
	}
}

func TestCountBeastStates(t *testing.T) {
	beasts := []*senju.Beast{
		{HP: 10, State: senju.Idle},
		{HP: 5, State: senju.Patrolling},
		{HP: 0, State: senju.Stunned},
		{HP: 3, State: senju.Stunned},
	}

	alive, stunned := countBeastStates(beasts)
	if alive != 2 {
		t.Errorf("alive: got %d, want 2", alive)
	}
	if stunned != 2 {
		t.Errorf("stunned: got %d, want 2", stunned)
	}
}

func TestCountInvasionState(t *testing.T) {
	waves := []*invasion.InvasionWave{
		{
			State: invasion.Active,
			Invaders: []*invasion.Invader{
				{State: invasion.Advancing},
				{State: invasion.Fighting},
				{State: invasion.Defeated},
			},
		},
		{
			State: invasion.Completed,
			Invaders: []*invasion.Invader{
				{State: invasion.Defeated},
			},
		},
	}

	active, completed, total, activeInv := countInvasionState(waves)
	if active != 1 {
		t.Errorf("active waves: got %d, want 1", active)
	}
	if completed != 1 {
		t.Errorf("completed waves: got %d, want 1", completed)
	}
	if total != 4 {
		t.Errorf("total invaders: got %d, want 4", total)
	}
	if activeInv != 2 {
		t.Errorf("active invaders: got %d, want 2", activeInv)
	}
}
