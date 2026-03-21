package simulation

import (
	"testing"

	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/types"
)

func TestCreateCheckpoint_RequiresCheckpointableRNG(t *testing.T) {
	sc := minimalScenario()
	// Use a plain seededRNG (not checkpointable).
	rng := types.NewSeededRNG(1)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	_, err = CreateCheckpoint(engine)
	if err == nil {
		t.Fatal("expected error when RNG is not CheckpointableRNG")
	}
}

func TestCreateCheckpoint_Success(t *testing.T) {
	sc := minimalScenario()
	rng := types.NewCheckpointableRNG(42)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	// Run a few ticks to advance state.
	for i := range 3 {
		if _, err := engine.Step([]PlayerAction{NoAction{}}); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}

	cp, err := CreateCheckpoint(engine)
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	if len(cp.CaveData) == 0 {
		t.Error("CaveData is empty")
	}
	if len(cp.ChiFlowData) == 0 {
		t.Error("ChiFlowData is empty")
	}
	if len(cp.EconomyData) == 0 {
		t.Error("EconomyData is empty")
	}
	if len(cp.ProgressData) == 0 {
		t.Error("ProgressData is empty")
	}
	if cp.RNGState.Seed != 42 {
		t.Errorf("RNGState.Seed = %d, want 42", cp.RNGState.Seed)
	}
	if cp.RNGState.Calls < 0 {
		t.Errorf("RNGState.Calls = %d, want >= 0", cp.RNGState.Calls)
	}
}

func TestCheckpoint_RestoreThenContinue_SameResult(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.DragonVeins = []scenario.DragonVeinPlacement{
		{SourcePos: types.Pos{X: 6, Y: 6}, Element: types.Wood, FlowRate: 5.0},
	}
	sc.InitialState.StartingBeasts = []scenario.BeastPlacement{
		{SpeciesID: "suiryu", RoomIndex: 0},
	}

	const seed int64 = 99
	const ticksBeforeCheckpoint = 5
	const ticksAfterCheckpoint = 10

	// --- Path A: run full simulation ---
	engineA, err := NewSimulationEngine(sc, types.NewCheckpointableRNG(seed))
	if err != nil {
		t.Fatalf("NewSimulationEngine (A): %v", err)
	}

	// Run to checkpoint point.
	for i := range ticksBeforeCheckpoint {
		if _, err := engineA.Step([]PlayerAction{NoAction{}}); err != nil {
			t.Fatalf("Step (A pre-cp) %d: %v", i, err)
		}
	}

	// Create checkpoint.
	cp, err := CreateCheckpoint(engineA)
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	// Continue running path A.
	var resultA GameResult
	for i := range ticksAfterCheckpoint {
		resultA, err = engineA.Step([]PlayerAction{NoAction{}})
		if err != nil {
			t.Fatalf("Step (A post-cp) %d: %v", i, err)
		}
	}
	tickA := engineA.State.Progress.CurrentTick
	balanceA := engineA.State.EconomyEngine.ChiPool.Balance()
	coreHPA := engineA.State.Progress.CoreHP
	beastCountA := len(engineA.State.Beasts)

	// --- Path B: restore from checkpoint and continue ---
	engineB, err := RestoreCheckpoint(cp, sc)
	if err != nil {
		t.Fatalf("RestoreCheckpoint: %v", err)
	}

	// Continue running path B with same actions.
	var resultB GameResult
	for i := range ticksAfterCheckpoint {
		resultB, err = engineB.Step([]PlayerAction{NoAction{}})
		if err != nil {
			t.Fatalf("Step (B post-cp) %d: %v", i, err)
		}
	}
	tickB := engineB.State.Progress.CurrentTick
	balanceB := engineB.State.EconomyEngine.ChiPool.Balance()
	coreHPB := engineB.State.Progress.CoreHP
	beastCountB := len(engineB.State.Beasts)

	// --- Verify both paths produce identical results ---
	if tickA != tickB {
		t.Errorf("tick mismatch: A=%d, B=%d", tickA, tickB)
	}
	if balanceA != balanceB {
		t.Errorf("chi balance mismatch: A=%.4f, B=%.4f", balanceA, balanceB)
	}
	if coreHPA != coreHPB {
		t.Errorf("core HP mismatch: A=%d, B=%d", coreHPA, coreHPB)
	}
	if beastCountA != beastCountB {
		t.Errorf("beast count mismatch: A=%d, B=%d", beastCountA, beastCountB)
	}
	if resultA.Status != resultB.Status {
		t.Errorf("result status mismatch: A=%v, B=%v", resultA.Status, resultB.Status)
	}
}

func TestCheckpoint_RestorePreservesState(t *testing.T) {
	sc := minimalScenario()
	sc.InitialState.StartingBeasts = []scenario.BeastPlacement{
		{SpeciesID: "suiryu", RoomIndex: 0},
	}
	rng := types.NewCheckpointableRNG(7)

	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	// Run a few ticks.
	for i := range 5 {
		if _, err := engine.Step([]PlayerAction{NoAction{}}); err != nil {
			t.Fatalf("Step %d: %v", i, err)
		}
	}

	// Capture expected values.
	wantTick := engine.State.Progress.CurrentTick
	wantBalance := engine.State.EconomyEngine.ChiPool.Balance()
	wantCoreHP := engine.State.Progress.CoreHP
	wantBeastCount := len(engine.State.Beasts)
	wantNextBeastID := engine.State.NextBeastID
	wantNextWaveID := engine.State.NextWaveID

	cp, err := CreateCheckpoint(engine)
	if err != nil {
		t.Fatalf("CreateCheckpoint: %v", err)
	}

	restored, err := RestoreCheckpoint(cp, sc)
	if err != nil {
		t.Fatalf("RestoreCheckpoint: %v", err)
	}

	rs := restored.State
	if rs.Progress.CurrentTick != wantTick {
		t.Errorf("CurrentTick = %d, want %d", rs.Progress.CurrentTick, wantTick)
	}
	if rs.EconomyEngine.ChiPool.Balance() != wantBalance {
		t.Errorf("chi balance = %.4f, want %.4f", rs.EconomyEngine.ChiPool.Balance(), wantBalance)
	}
	if rs.Progress.CoreHP != wantCoreHP {
		t.Errorf("CoreHP = %d, want %d", rs.Progress.CoreHP, wantCoreHP)
	}
	if len(rs.Beasts) != wantBeastCount {
		t.Errorf("beast count = %d, want %d", len(rs.Beasts), wantBeastCount)
	}
	if rs.NextBeastID != wantNextBeastID {
		t.Errorf("NextBeastID = %d, want %d", rs.NextBeastID, wantNextBeastID)
	}
	if rs.NextWaveID != wantNextWaveID {
		t.Errorf("NextWaveID = %d, want %d", rs.NextWaveID, wantNextWaveID)
	}

	// Verify subsystem engines are not nil.
	if rs.Cave == nil {
		t.Error("restored Cave is nil")
	}
	if rs.ChiFlowEngine == nil {
		t.Error("restored ChiFlowEngine is nil")
	}
	if rs.GrowthEngine == nil {
		t.Error("restored GrowthEngine is nil")
	}
	if rs.BehaviorEngine == nil {
		t.Error("restored BehaviorEngine is nil")
	}
	if rs.InvasionEngine == nil {
		t.Error("restored InvasionEngine is nil")
	}
	if rs.EconomyEngine == nil {
		t.Error("restored EconomyEngine is nil")
	}
	if rs.EventEngine == nil {
		t.Error("restored EventEngine is nil")
	}
}
