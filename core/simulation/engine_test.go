package simulation

import (
	"encoding/json"
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
)

// minimalScenario returns a Scenario with the minimum required configuration
// for NewSimulationEngine to succeed.
func minimalScenario() *scenario.Scenario {
	return &scenario.Scenario{
		ID:         "test_scenario",
		Name:       "Test",
		Difficulty: "easy",
		InitialState: scenario.InitialState{
			CaveWidth:      20,
			CaveHeight:     20,
			TerrainSeed:    42,
			TerrainDensity: 0.0,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
			},
			StartingChi: 100.0,
		},
	}
}

func TestNewSimulationEngine_Minimal(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	if engine.State == nil {
		t.Fatal("State is nil")
	}
	if engine.Executor == nil {
		t.Fatal("Executor is nil")
	}
	if engine.State.Cave == nil {
		t.Fatal("Cave is nil")
	}
	if engine.State.RNG == nil {
		t.Fatal("RNG is nil")
	}
}

func TestNewSimulationEngine_CaveSize(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	grid := engine.State.Cave.Grid
	if grid.Width != 20 || grid.Height != 20 {
		t.Errorf("cave size = %dx%d, want 20x20", grid.Width, grid.Height)
	}
}

func TestNewSimulationEngine_PrebuiltRoom(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	rooms := engine.State.Cave.Rooms
	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}
	room := rooms[0]
	if room.TypeID != "dragon_hole" {
		t.Errorf("room TypeID = %q, want %q", room.TypeID, "dragon_hole")
	}
	if room.Level != 1 {
		t.Errorf("room Level = %d, want 1", room.Level)
	}
	if room.CoreHP <= 0 {
		t.Errorf("dragon_hole CoreHP = %d, want > 0", room.CoreHP)
	}
}

func TestNewSimulationEngine_CoreHP(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	// Progress CoreHP should match the dragon hole room's CoreHP.
	room := engine.State.Cave.Rooms[0]
	if engine.State.Progress.CoreHP != room.CoreHP {
		t.Errorf("Progress CoreHP = %d, want %d", engine.State.Progress.CoreHP, room.CoreHP)
	}
}

func TestNewSimulationEngine_ChiPool(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	balance := engine.State.EconomyEngine.ChiPool.Balance()
	if balance != 100.0 {
		t.Errorf("chi balance = %.1f, want 100.0", balance)
	}
}

func TestNewSimulationEngine_Registries(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	if engine.State.RoomTypeRegistry == nil {
		t.Error("RoomTypeRegistry is nil")
	}
	if engine.State.SpeciesRegistry == nil {
		t.Error("SpeciesRegistry is nil")
	}
	if engine.State.EvolutionRegistry == nil {
		t.Error("EvolutionRegistry is nil")
	}
	if engine.State.InvaderClassRegistry == nil {
		t.Error("InvaderClassRegistry is nil")
	}
}

func TestNewSimulationEngine_SubsystemEngines(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	s := engine.State
	if s.ChiFlowEngine == nil {
		t.Error("ChiFlowEngine is nil")
	}
	if s.GrowthEngine == nil {
		t.Error("GrowthEngine is nil")
	}
	if s.BehaviorEngine == nil {
		t.Error("BehaviorEngine is nil")
	}
	if s.DefeatProcessor == nil {
		t.Error("DefeatProcessor is nil")
	}
	if s.InvasionEngine == nil {
		t.Error("InvasionEngine is nil")
	}
	if s.EconomyEngine == nil {
		t.Error("EconomyEngine is nil")
	}
	if s.EventEngine == nil {
		t.Error("EventEngine is nil")
	}
}

func TestNewSimulationEngine_Progress(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	p := engine.State.Progress
	if p.ScenarioID != "test_scenario" {
		t.Errorf("ScenarioID = %q, want %q", p.ScenarioID, "test_scenario")
	}
	if p.CurrentTick != 0 {
		t.Errorf("CurrentTick = %d, want 0", p.CurrentTick)
	}
}

func TestNewSimulationEngine_WithDragonVeins(t *testing.T) {
	sc := minimalScenario()
	// Place the dragon vein source inside the dragon_hole room area.
	sc.InitialState.DragonVeins = []scenario.DragonVeinPlacement{
		{SourcePos: types.Pos{X: 6, Y: 6}, Element: types.Wood, FlowRate: 5.0},
	}
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	if engine.State.ChiFlowEngine == nil {
		t.Fatal("ChiFlowEngine is nil")
	}
}

func TestNewSimulationEngine_WithStartingBeasts(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.StartingBeasts = []scenario.BeastPlacement{
		{SpeciesID: "suiryu", RoomIndex: 0},
	}
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	if len(engine.State.Beasts) != 1 {
		t.Fatalf("expected 1 beast, got %d", len(engine.State.Beasts))
	}
	beast := engine.State.Beasts[0]
	if beast.ID != 1 {
		t.Errorf("beast ID = %d, want 1", beast.ID)
	}
	if engine.State.NextBeastID != 2 {
		t.Errorf("NextBeastID = %d, want 2", engine.State.NextBeastID)
	}
}

func TestNewSimulationEngine_InvalidRoomType(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.PrebuiltRooms = []scenario.RoomPlacement{
		{TypeID: "nonexistent_room", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
	}
	rng := types.NewSeededRNG(1)

	_, err := NewSimulationEngine(sc, rng)
	if err == nil {
		t.Fatal("expected error for invalid room type")
	}
}

func TestNewSimulationEngine_InvalidBeastSpecies(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.StartingBeasts = []scenario.BeastPlacement{
		{SpeciesID: "nonexistent_species", RoomIndex: -1},
	}
	rng := types.NewSeededRNG(1)

	_, err := NewSimulationEngine(sc, rng)
	if err == nil {
		t.Fatal("expected error for invalid beast species")
	}
}

func TestNewSimulationEngine_DefaultLevel(t *testing.T) {
	sc := minimalScenario()
	// Level 0 should default to 1.
	sc.InitialState.PrebuiltRooms[0].Level = 0
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	room := engine.State.Cave.Rooms[0]
	if room.Level != 1 {
		t.Errorf("room Level = %d, want 1 (default)", room.Level)
	}
}

func TestRun_SingleTick(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	result, err := engine.Run(1, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// With no win/lose conditions and only 1 tick, maxTicks is reached → Lost.
	if result.Status != Lost {
		t.Errorf("Status = %v, want Lost (max ticks reached)", result.Status)
	}
	if result.Reason != "max ticks reached" {
		t.Errorf("Reason = %q, want %q", result.Reason, "max ticks reached")
	}
	// One tick should have been executed.
	if engine.State.Progress.CurrentTick != 1 {
		t.Errorf("CurrentTick = %d, want 1", engine.State.Progress.CurrentTick)
	}
	if len(engine.TickLog) != 1 {
		t.Errorf("TickLog length = %d, want 1", len(engine.TickLog))
	}
}

func TestRun_WinCondition(t *testing.T) {
	sc := minimalScenario()
	sc.WinConditions = []scenario.ConditionDef{
		{Type: "survive_until", Params: json.RawMessage(`{"ticks": 5}`)},
	}
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	result, err := engine.Run(100, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != Won {
		t.Errorf("Status = %v, want Won", result.Status)
	}
	if result.FinalTick > 10 {
		t.Errorf("FinalTick = %d, expected <= 10", result.FinalTick)
	}
}

func TestRun_LoseCondition(t *testing.T) {
	sc := minimalScenario()
	sc.LoseConditions = []scenario.ConditionDef{
		{Type: "core_destroyed"},
	}
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	// Set core HP to 0 to trigger immediate loss.
	engine.State.Progress.CoreHP = 0

	result, err := engine.Run(100, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != Lost {
		t.Errorf("Status = %v, want Lost", result.Status)
	}
	if result.FinalTick != 0 {
		t.Errorf("FinalTick = %d, want 0", result.FinalTick)
	}
}

func TestRun_PlayerAction(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	callCount := 0
	provider := func(snap scenario.GameSnapshot) []PlayerAction {
		callCount++
		return []PlayerAction{NoAction{}}
	}

	result, err := engine.Run(3, provider)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != Lost {
		t.Errorf("Status = %v, want Lost (max ticks)", result.Status)
	}
	if callCount != 3 {
		t.Errorf("actionProvider called %d times, want 3", callCount)
	}
	if engine.State.Progress.CurrentTick != 3 {
		t.Errorf("CurrentTick = %d, want 3", engine.State.Progress.CurrentTick)
	}
}

func TestRun_MaxTicksLimit(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	result, err := engine.Run(10, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Status != Lost {
		t.Errorf("Status = %v, want Lost", result.Status)
	}
	if result.Reason != "max ticks reached" {
		t.Errorf("Reason = %q, want %q", result.Reason, "max ticks reached")
	}
	if engine.State.Progress.CurrentTick != 10 {
		t.Errorf("CurrentTick = %d, want 10", engine.State.Progress.CurrentTick)
	}
	if len(engine.TickLog) != 10 {
		t.Errorf("TickLog length = %d, want 10", len(engine.TickLog))
	}
}

func TestNormalizeCaveScore_InRange(t *testing.T) {
	params := fengshui.DefaultScoreParams()
	tests := []struct {
		name     string
		raw      float64
		numRooms int
		wantMin  float64
		wantMax  float64
	}{
		{"zero rooms", 100.0, 0, 0.0, 0.0},
		{"zero score", 0.0, 5, 0.0, 0.0},
		{"half score", 525.0, 5, 0.0, 1.0}, // 525 / (5*210) = 0.5
		{"max score", 1050.0, 5, 1.0, 1.0}, // 1050 / (5*210) = 1.0
		{"over max", 2000.0, 5, 1.0, 1.0},  // clamped to 1.0
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeCaveScore(tt.raw, tt.numRooms, params)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("normalizeCaveScore(%v, %d) = %v, want [%v, %v]",
					tt.raw, tt.numRooms, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestNormalizeCaveScore_Precise(t *testing.T) {
	params := fengshui.DefaultScoreParams()
	// MaxRoomScore(4) = 100 + 30 + 4*20 = 210
	// 5 rooms → maxPossible = 1050
	got := normalizeCaveScore(525.0, 5, params)
	want := 0.5
	if got != want {
		t.Errorf("normalizeCaveScore(525, 5) = %v, want %v", got, want)
	}
}

func TestNormalizeCaveScore_SupplySublinear(t *testing.T) {
	// Verify that supply does not grow exponentially with room count.
	// With normalization, caveScore stays in [0,1], so supply is bounded.
	params := fengshui.DefaultScoreParams()

	// Simulate increasing room counts with proportional raw scores.
	// Each room contributes ~100 raw score (moderate feng shui).
	supplyForRooms := func(numRooms int) float64 {
		rawScore := float64(numRooms) * 100.0
		return normalizeCaveScore(rawScore, numRooms, params)
	}

	score1 := supplyForRooms(1)
	score10 := supplyForRooms(10)
	score100 := supplyForRooms(100)

	// All should be the same since raw/room is constant.
	if score1 != score10 || score10 != score100 {
		t.Errorf("normalized scores differ with constant per-room score: 1=%v, 10=%v, 100=%v",
			score1, score10, score100)
	}

	// And all should be ~100/210 ≈ 0.476
	want := 100.0 / 210.0
	if score1 < want-0.001 || score1 > want+0.001 {
		t.Errorf("normalized score = %v, want ~%v", score1, want)
	}
}

func TestNewSimulationEngine_Deterministic(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.DragonVeins = []scenario.DragonVeinPlacement{
		{SourcePos: types.Pos{X: 6, Y: 6}, Element: types.Fire, FlowRate: 3.0},
	}

	e1, err := NewSimulationEngine(sc, types.NewSeededRNG(99))
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	e2, err := NewSimulationEngine(sc, types.NewSeededRNG(99))
	if err != nil {
		t.Fatalf("second: %v", err)
	}

	// Same scenario + same RNG seed → same CoreHP.
	if e1.State.Progress.CoreHP != e2.State.Progress.CoreHP {
		t.Errorf("CoreHP mismatch: %d vs %d", e1.State.Progress.CoreHP, e2.State.Progress.CoreHP)
	}
	// Same chi balance.
	b1 := e1.State.EconomyEngine.ChiPool.Balance()
	b2 := e2.State.EconomyEngine.ChiPool.Balance()
	if b1 != b2 {
		t.Errorf("chi balance mismatch: %.1f vs %.1f", b1, b2)
	}
}
