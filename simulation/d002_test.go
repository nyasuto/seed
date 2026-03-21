package simulation

import (
	"fmt"
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
)

// d002StandardScenario returns a scenario suitable for D002 validation.
// It includes terrain density, invasion waves, and resource constraints to
// exercise all three D002 principles.
func d002StandardScenario(terrainSeed int64) *scenario.Scenario {
	return &scenario.Scenario{
		ID:         "d002_standard",
		Name:       "D002 Standard",
		Difficulty: "normal",
		InitialState: scenario.InitialState{
			CaveWidth:      30,
			CaveHeight:     30,
			TerrainSeed:    terrainSeed,
			TerrainDensity: 0.3,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 10, Y: 10}, Level: 1},
			},
			DragonVeins: []scenario.DragonVeinPlacement{
				{SourcePos: types.Pos{X: 10, Y: 12}, Element: types.Wood, FlowRate: 5.0},
				{SourcePos: types.Pos{X: 12, Y: 10}, Element: types.Fire, FlowRate: 3.0},
			},
			StartingChi: 300.0,
		},
		WinConditions: []scenario.ConditionDef{
			{Type: "survive_until", Params: map[string]any{"ticks": 200.0}},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed"},
			{Type: "bankrupt", Params: map[string]any{"ticks": 20.0}},
		},
		WaveSchedule: []scenario.WaveScheduleEntry{
			{TriggerTick: 40, Difficulty: 1.0, MinInvaders: 2, MaxInvaders: 4},
			{TriggerTick: 80, Difficulty: 1.2, MinInvaders: 3, MaxInvaders: 5},
			{TriggerTick: 130, Difficulty: 1.5, MinInvaders: 3, MaxInvaders: 6},
			{TriggerTick: 170, Difficulty: 1.8, MinInvaders: 4, MaxInvaders: 7},
		},
		Constraints: scenario.GameConstraints{
			MaxRooms:  15,
			MaxBeasts: 8,
		},
	}
}

// TestD002_Principle1_ImperfectionForced verifies D002 principle 1:
// terrain randomness forces different room placements across different seeds,
// meaning no single optimal layout exists.
//
// Verification:
//   - Run SimpleAI on the same scenario with 10 different terrain seeds.
//   - Record room count, positions, and feng shui scores for each run.
//   - Assert that layouts differ: not all runs produce the same room count
//     and room positions, proving terrain constrains placement differently.
func TestD002_Principle1_ImperfectionForced(t *testing.T) {
	const numSeeds = 10

	type runResult struct {
		terrainSeed   int64
		fengShuiScore float64
		roomCount     int
		roomPositions []types.Pos
		gameStatus    GameStatus
	}

	results := make([]runResult, numSeeds)

	for i := 0; i < numSeeds; i++ {
		terrainSeed := int64(i*1000 + 1)
		sc := d002StandardScenario(terrainSeed)
		rng := types.NewSeededRNG(42) // same engine RNG for all runs

		engine, err := NewSimulationEngine(sc, rng)
		if err != nil {
			t.Fatalf("seed %d: NewSimulationEngine: %v", terrainSeed, err)
		}

		ai := NewSimpleAIPlayer(engine.State)
		result, err := engine.Run(300, func(snap scenario.GameSnapshot) []PlayerAction {
			return ai.DecideActions(snap)
		})
		if err != nil {
			t.Fatalf("seed %d: Run: %v", terrainSeed, err)
		}

		// Calculate final feng shui score.
		ev := fengshui.NewEvaluator(
			engine.State.Cave,
			engine.State.RoomTypeRegistry,
			engine.State.ScoreParams,
		)
		score := ev.CaveTotal(engine.State.ChiFlowEngine)

		var positions []types.Pos
		for _, room := range engine.State.Cave.Rooms {
			positions = append(positions, room.Pos)
		}

		results[i] = runResult{
			terrainSeed:   terrainSeed,
			fengShuiScore: score,
			roomCount:     len(engine.State.Cave.Rooms),
			roomPositions: positions,
			gameStatus:    result.Status,
		}

		t.Logf("terrain_seed=%d: status=%v rooms=%d fengshui=%.2f",
			terrainSeed, result.Status, len(engine.State.Cave.Rooms), score)
	}

	// Assertion 1: room counts must not all be identical.
	// Different terrain restricts buildable area differently, so the number of
	// rooms the AI manages to build should vary.
	allSameCount := true
	firstCount := results[0].roomCount
	for _, r := range results[1:] {
		if r.roomCount != firstCount {
			allSameCount = false
			break
		}
	}
	if allSameCount {
		t.Errorf("all %d terrain seeds produced identical room count %d; "+
			"terrain randomness should affect buildable area", numSeeds, firstCount)
	}

	// Assertion 2: room position layouts must not all be identical.
	// Even if room counts coincide, positions should differ because terrain
	// blocks different cells.
	allSamePositions := true
	for i := 1; i < len(results); i++ {
		if !positionsEqual(results[0].roomPositions, results[i].roomPositions) {
			allSamePositions = false
			break
		}
	}
	if allSamePositions {
		t.Errorf("all %d terrain seeds produced identical room positions; "+
			"terrain randomness should force different placements", numSeeds)
	}

	// Count distinct layout signatures to measure diversity.
	signatures := make(map[string]int)
	for _, r := range results {
		sig := fmt.Sprintf("%d:%v", r.roomCount, r.roomPositions)
		signatures[sig]++
	}
	distinctLayouts := len(signatures)
	t.Logf("distinct layouts: %d / %d seeds", distinctLayouts, numSeeds)

	// At least 3 distinct layouts should emerge from 10 different terrain seeds.
	if distinctLayouts < 3 {
		t.Errorf("only %d distinct layouts from %d terrain seeds; "+
			"expected at least 3 to demonstrate meaningful terrain influence",
			distinctLayouts, numSeeds)
	}

	// Log score range (informational — scores may be identical if SimpleAI
	// does not build corridors, limiting chi propagation to new rooms).
	minScore, maxScore := results[0].fengShuiScore, results[0].fengShuiScore
	for _, r := range results[1:] {
		if r.fengShuiScore < minScore {
			minScore = r.fengShuiScore
		}
		if r.fengShuiScore > maxScore {
			maxScore = r.fengShuiScore
		}
	}
	t.Logf("feng shui score range: %.2f – %.2f (spread: %.2f)", minScore, maxScore, maxScore-minScore)
}

// positionsEqual checks if two position slices are identical (same order, same values).
func positionsEqual(a, b []types.Pos) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
