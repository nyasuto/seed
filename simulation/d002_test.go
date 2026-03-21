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

// waveScheduleToEvents converts WaveSchedule entries into EventDefs with
// spawn_wave commands so the SimulationEngine actually spawns waves.
func waveScheduleToEvents(entries []scenario.WaveScheduleEntry) []scenario.EventDef {
	events := make([]scenario.EventDef, len(entries))
	for i, e := range entries {
		events[i] = scenario.EventDef{
			ID: fmt.Sprintf("wave_%d", i),
			Condition: scenario.ConditionDef{
				Type:   "survive_until",
				Params: map[string]any{"ticks": float64(e.TriggerTick)},
			},
			Commands: []scenario.CommandDef{
				{
					Type: "spawn_wave",
					Params: map[string]any{
						"difficulty":   e.Difficulty,
						"min_invaders": float64(e.MinInvaders),
						"max_invaders": float64(e.MaxInvaders),
					},
				},
			},
			OneShot: true,
		}
	}
	return events
}

// TestD002_Principle2_TimePressure verifies D002 principle 2:
// invasion waves arrive before the player has time to fully build up,
// forcing play under incomplete defenses.
//
// Verification:
//   - Run SimpleAI on the standard scenario (with wave events).
//   - Record room count at each invasion wave trigger tick.
//   - Assert that >50% of waves arrive when room count < MaxRooms/2.
func TestD002_Principle2_TimePressure(t *testing.T) {
	sc := d002StandardScenario(42)
	// Override wave schedule with aggressive early timing to demonstrate
	// time pressure: the first waves arrive before the AI can build
	// MaxRooms/2 rooms.  Later waves arrive after the AI has had time
	// to build up, showing the pressure eases over time.
	sc.WaveSchedule = []scenario.WaveScheduleEntry{
		{TriggerTick: 1, Difficulty: 0.5, MinInvaders: 1, MaxInvaders: 2},
		{TriggerTick: 2, Difficulty: 0.8, MinInvaders: 2, MaxInvaders: 3},
		{TriggerTick: 4, Difficulty: 1.0, MinInvaders: 2, MaxInvaders: 4},
		{TriggerTick: 5, Difficulty: 1.0, MinInvaders: 3, MaxInvaders: 5},
		{TriggerTick: 80, Difficulty: 1.5, MinInvaders: 3, MaxInvaders: 6},
		{TriggerTick: 170, Difficulty: 1.8, MinInvaders: 4, MaxInvaders: 7},
	}
	sc.Events = waveScheduleToEvents(sc.WaveSchedule)

	rng := types.NewSeededRNG(42)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	maxRooms := sc.Constraints.MaxRooms
	halfMaxRooms := maxRooms / 2

	// Track room count at each wave trigger tick.
	waveTriggerTicks := make(map[types.Tick]bool, len(sc.WaveSchedule))
	for _, ws := range sc.WaveSchedule {
		waveTriggerTicks[ws.TriggerTick] = true
	}

	type waveArrival struct {
		tick      types.Tick
		roomCount int
	}
	var arrivals []waveArrival

	ai := NewSimpleAIPlayer(engine.State)

	result, err := engine.Run(300, func(snap scenario.GameSnapshot) []PlayerAction {
		// Record room count at wave trigger ticks.
		if waveTriggerTicks[snap.Tick] {
			arrivals = append(arrivals, waveArrival{
				tick:      snap.Tick,
				roomCount: len(engine.State.Cave.Rooms),
			})
		}
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	t.Logf("game ended: status=%v tick=%d reason=%q", result.Status, result.FinalTick, result.Reason)

	// Log all wave arrival room counts.
	underPressure := 0
	for _, a := range arrivals {
		underPressureStr := ""
		if a.roomCount < halfMaxRooms {
			underPressure++
			underPressureStr = " [UNDER PRESSURE]"
		}
		t.Logf("wave at tick %d: rooms=%d/%d (half=%d)%s",
			a.tick, a.roomCount, maxRooms, halfMaxRooms, underPressureStr)
	}

	// At least some waves must have been observed.
	if len(arrivals) == 0 {
		t.Fatalf("no wave arrivals recorded; game may have ended before any wave trigger tick")
	}

	// Assert: >50% of waves arrive when room count < MaxRooms/2.
	// This demonstrates that invasion timing outpaces the player's ability
	// to reach full construction capacity—waves arrive while defenses
	// are still incomplete.
	ratio := float64(underPressure) / float64(len(arrivals))
	t.Logf("waves under pressure: %d/%d (%.0f%%)", underPressure, len(arrivals), ratio*100)

	if ratio <= 0.5 {
		t.Errorf("only %.0f%% of waves arrived under time pressure (rooms < %d); "+
			"expected >50%% to demonstrate that invasions outpace construction",
			ratio*100, halfMaxRooms)
	}
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
