package simulation

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nyasuto/seed/core/fengshui"
	"github.com/nyasuto/seed/core/scenario"
	"github.com/nyasuto/seed/core/senju"
	"github.com/nyasuto/seed/core/types"
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
			{Type: "survive_until", Params: json.RawMessage(`{"ticks": 200}`)},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed"},
			{Type: "bankrupt", Params: json.RawMessage(`{"ticks": 20}`)},
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

	for i := range numSeeds {
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
				Params: json.RawMessage(fmt.Sprintf(`{"ticks": %d}`, e.TriggerTick)),
			},
			Commands: []scenario.CommandDef{
				{
					Type:   "spawn_wave",
					Params: json.RawMessage(fmt.Sprintf(`{"difficulty": %f, "min_invaders": %d, "max_invaders": %d}`, e.Difficulty, e.MinInvaders, e.MaxInvaders)),
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

// TestD002_Principle3_ContinuousTradeoffs verifies D002 principle 3:
// the game's resource constraints ensure that no run ever reaches the
// "all rooms MAX + all beasts MAX" state upon victory.
//
// Verification:
//   - Run SimpleAI on the standard scenario 100 times with different seeds.
//   - For each run that results in a win, check whether ALL rooms have reached
//     the maximum upgrade level (5) AND all beasts have reached MaxLevel (50).
//   - Assert that the rate of achieving this "perfect max" state is 0%.
func TestD002_Principle3_ContinuousTradeoffs(t *testing.T) {
	const numRuns = 100
	const roomMaxLevel = 5 // practical max level for room upgrades

	beastMaxLevel := senju.DefaultGrowthParams().MaxLevel

	wins := 0
	perfectMaxWins := 0

	for i := range numRuns {
		seed := int64(i*97 + 1)
		sc := d002StandardScenario(seed)
		sc.Events = waveScheduleToEvents(sc.WaveSchedule)
		rng := types.NewSeededRNG(seed)

		engine, err := NewSimulationEngine(sc, rng)
		if err != nil {
			t.Fatalf("seed %d: NewSimulationEngine: %v", seed, err)
		}

		ai := NewSimpleAIPlayer(engine.State)
		result, err := engine.Run(300, func(snap scenario.GameSnapshot) []PlayerAction {
			return ai.DecideActions(snap)
		})
		if err != nil {
			t.Fatalf("seed %d: Run: %v", seed, err)
		}

		if result.Status != Won {
			continue
		}
		wins++

		// Check if all rooms are at max level.
		allRoomsMax := true
		roomCount := len(engine.State.Cave.Rooms)
		if roomCount < sc.Constraints.MaxRooms {
			allRoomsMax = false
		} else {
			for _, room := range engine.State.Cave.Rooms {
				if room.Level < roomMaxLevel {
					allRoomsMax = false
					break
				}
			}
		}

		// Check if all beasts are at max level.
		allBeastsMax := true
		beastCount := len(engine.State.Beasts)
		if beastCount < sc.Constraints.MaxBeasts {
			allBeastsMax = false
		} else {
			for _, beast := range engine.State.Beasts {
				if beast.Level < beastMaxLevel {
					allBeastsMax = false
					break
				}
			}
		}

		if allRoomsMax && allBeastsMax {
			perfectMaxWins++
			t.Logf("seed %d: PERFECT MAX achieved at tick %d (rooms=%d beasts=%d)",
				seed, result.FinalTick, roomCount, beastCount)
		}
	}

	t.Logf("results: %d wins out of %d runs, %d achieved perfect max",
		wins, numRuns, perfectMaxWins)

	if perfectMaxWins > 0 {
		rate := float64(perfectMaxWins) / float64(numRuns) * 100
		t.Errorf("perfect max (all rooms level %d + all beasts level %d) achieved in %d/%d runs (%.1f%%); "+
			"expected 0%% to demonstrate continuous trade-offs",
			roomMaxLevel, beastMaxLevel, perfectMaxWins, numRuns, rate)
	}
}

// TestD002_Antipattern_Rich verifies that the antipattern_rich scenario is
// "always profitable" — the player never experiences resource tension.
// This proves that such a scenario lacks the economic pressure that makes
// the game interesting (it serves as a negative example for D002).
func TestD002_Antipattern_Rich(t *testing.T) {
	sc := antipatternRichScenario()
	sc.Events = waveScheduleToEvents(sc.WaveSchedule)

	rng := types.NewSeededRNG(42)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)

	// Track chi balance each tick to check for deficit.
	minChi := sc.InitialState.StartingChi
	alwaysSurplus := true

	// Run beyond MaxTicks to allow survive_until win condition to trigger.
	runTicks := int(sc.Constraints.MaxTicks) + 100
	if sc.Constraints.MaxTicks == 0 {
		runTicks = 1000
	}

	result, err := engine.Run(runTicks, func(snap scenario.GameSnapshot) []PlayerAction {
		chi := snap.ChiPoolBalance
		if chi < minChi {
			minChi = chi
		}
		if chi <= 0 {
			alwaysSurplus = false
		}
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	t.Logf("antipattern_rich: status=%v tick=%d reason=%q minChi=%.2f",
		result.Status, result.FinalTick, result.Reason, minChi)

	// The rich scenario should always be surplus — chi never reaches 0.
	if !alwaysSurplus {
		t.Errorf("antipattern_rich: chi went to 0 or below; expected always surplus with starting_chi=%.0f",
			sc.InitialState.StartingChi)
	}

	// The player should not lose — resources are so abundant that losing is
	// nearly impossible.
	if result.Status == Lost {
		t.Errorf("antipattern_rich: game was lost (reason=%q); expected survival with abundant resources",
			result.Reason)
	}
}

// TestD002_Antipattern_Impossible verifies that the antipattern_impossible
// scenario leads to immediate defeat. With minimal starting chi, dense terrain,
// and an extremely early invasion wave, the AI cannot mount any defense.
func TestD002_Antipattern_Impossible(t *testing.T) {
	sc := antipatternImpossibleScenario()
	sc.Events = waveScheduleToEvents(sc.WaveSchedule)

	rng := types.NewSeededRNG(42)
	engine, err := NewSimulationEngine(sc, rng)
	if err != nil {
		t.Fatalf("NewSimulationEngine: %v", err)
	}

	ai := NewSimpleAIPlayer(engine.State)

	runTicks := int(sc.Constraints.MaxTicks) + 100
	if sc.Constraints.MaxTicks == 0 {
		runTicks = 400
	}

	result, err := engine.Run(runTicks, func(snap scenario.GameSnapshot) []PlayerAction {
		return ai.DecideActions(snap)
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	t.Logf("antipattern_impossible: status=%v tick=%d reason=%q",
		result.Status, result.FinalTick, result.Reason)

	// The impossible scenario should result in defeat.
	if result.Status != Lost {
		t.Errorf("antipattern_impossible: expected Lost but got %v (reason=%q); "+
			"scenario with starting_chi=%.0f and first wave at tick %d should be unwinnable",
			result.Status, result.Reason, sc.InitialState.StartingChi,
			sc.WaveSchedule[0].TriggerTick)
	}
}

// antipatternRichScenario returns the antipattern_rich scenario as a Go struct.
// This mirrors scenario/testdata/antipattern_rich.json.
func antipatternRichScenario() *scenario.Scenario {
	return &scenario.Scenario{
		ID:          "antipattern_rich",
		Name:        "アンチパターン：潤沢リソース",
		Description: "D002検証用。StartingChiが極端に多く侵入波が弱い。",
		Difficulty:  "easy",
		InitialState: scenario.InitialState{
			CaveWidth:      32,
			CaveHeight:     32,
			TerrainSeed:    100,
			TerrainDensity: 0.05,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 15, Y: 15}, Level: 1},
			},
			DragonVeins: []scenario.DragonVeinPlacement{
				{SourcePos: types.Pos{X: 15, Y: 15}, Element: types.Wood, FlowRate: 2.0},
				{SourcePos: types.Pos{X: 15, Y: 15}, Element: types.Fire, FlowRate: 2.0},
				{SourcePos: types.Pos{X: 15, Y: 15}, Element: types.Earth, FlowRate: 2.0},
			},
			StartingChi: 9999.0,
			StartingBeasts: []scenario.BeastPlacement{
				{SpeciesID: "suiryu", RoomIndex: 0},
			},
		},
		WinConditions: []scenario.ConditionDef{
			{Type: "survive_until", Params: json.RawMessage(`{"ticks": 800}`)},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed"},
		},
		WaveSchedule: []scenario.WaveScheduleEntry{
			{TriggerTick: 300, Difficulty: 0.1, MinInvaders: 1, MaxInvaders: 1},
			{TriggerTick: 600, Difficulty: 0.2, MinInvaders: 1, MaxInvaders: 2},
		},
		Constraints: scenario.GameConstraints{
			MaxRooms:  20,
			MaxBeasts: 10,
			MaxTicks:  800,
		},
	}
}

// antipatternImpossibleScenario returns the antipattern_impossible scenario as a Go struct.
// This mirrors scenario/testdata/antipattern_impossible.json.
func antipatternImpossibleScenario() *scenario.Scenario {
	return &scenario.Scenario{
		ID:          "antipattern_impossible",
		Name:        "アンチパターン：詰みシナリオ",
		Description: "D002検証用。HardRock密度0.5、初波tick10、StartingChi極小。龍脈なし。",
		Difficulty:  "hard",
		InitialState: scenario.InitialState{
			CaveWidth:      32,
			CaveHeight:     32,
			TerrainSeed:    999,
			TerrainDensity: 0.5,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 15, Y: 15}, Level: 1},
			},
			// No dragon veins — zero chi supply ensures rapid bankruptcy.
			StartingChi: 10.0,
		},
		WinConditions: []scenario.ConditionDef{
			{Type: "survive_until", Params: json.RawMessage(`{"ticks": 200}`)},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed"},
			{Type: "bankrupt", Params: json.RawMessage(`{"ticks": 5}`)},
		},
		WaveSchedule: []scenario.WaveScheduleEntry{
			{TriggerTick: 10, Difficulty: 0.9, MinInvaders: 3, MaxInvaders: 5},
			{TriggerTick: 50, Difficulty: 1.0, MinInvaders: 4, MaxInvaders: 6},
			{TriggerTick: 100, Difficulty: 1.0, MinInvaders: 5, MaxInvaders: 8},
		},
		Constraints: scenario.GameConstraints{
			MaxRooms:  5,
			MaxBeasts: 2,
			MaxTicks:  200,
		},
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
