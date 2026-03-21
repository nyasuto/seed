package simulation

import (
	"testing"

	"github.com/ponpoko/chaosseed-core/fengshui"
	"github.com/ponpoko/chaosseed-core/scenario"
	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// buildBenchCave creates a cave with the specified number of rooms arranged in
// a line, connected by corridors, with a single dragon vein. It returns the
// ChiFlowEngine wired up to the cave.
func buildBenchCave(b *testing.B, numRooms int) (*world.Cave, *fengshui.ChiFlowEngine) {
	b.Helper()

	// Cave wide enough to hold rooms side by side: each room is 3 wide + 1 gap.
	caveWidth := numRooms*4 + 2
	caveHeight := 8
	cave, err := world.NewCave(caveWidth, caveHeight)
	if err != nil {
		b.Fatalf("NewCave: %v", err)
	}

	reg, err := world.LoadDefaultRoomTypes()
	if err != nil {
		b.Fatalf("LoadDefaultRoomTypes: %v", err)
	}

	// Room types to cycle through.
	roomTypes := []string{"dragon_hole", "chi_chamber", "senju_room", "trap_room", "recovery_room", "storage"}

	for i := range numRooms {
		x := 1 + i*4
		typeID := roomTypes[i%len(roomTypes)]
		_, err := cave.AddRoom(typeID, types.Pos{X: x, Y: 2}, 3, 3, []world.RoomEntrance{
			{Pos: types.Pos{X: x + 2, Y: 3}, Dir: types.East},
			{Pos: types.Pos{X: x, Y: 3}, Dir: types.West},
		})
		if err != nil {
			b.Fatalf("AddRoom %d: %v", i, err)
		}
	}

	// Connect rooms in a chain.
	for i := 1; i < numRooms; i++ {
		_, err := cave.ConnectRooms(i, i+1)
		if err != nil {
			b.Fatalf("ConnectRooms %d-%d: %v", i, i+1, err)
		}
	}

	// Build a dragon vein from the first room's position.
	source := types.Pos{X: 2, Y: 3}
	vein, err := fengshui.BuildDragonVein(cave, source, types.Wood, 10.0)
	if err != nil {
		b.Fatalf("BuildDragonVein: %v", err)
	}
	vein.ID = 1

	engine := fengshui.NewChiFlowEngine(cave, []*fengshui.DragonVein{vein}, reg, fengshui.DefaultFlowParams())
	return cave, engine
}

// BenchmarkOnCaveChanged measures the time to rebuild dragon veins after a cave
// change, for varying numbers of rooms.
func BenchmarkOnCaveChanged(b *testing.B) {
	cases := []struct {
		name     string
		numRooms int
	}{
		{"Rooms5", 5},
		{"Rooms10", 10},
		{"Rooms20", 20},
		{"Rooms50", 50},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			cave, engine := buildBenchCave(b, tc.numRooms)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.OnCaveChanged(cave)
			}
		})
	}
}

// BenchmarkFullTick measures the execution time of 100 ticks of a standard
// scenario with no player actions.
func BenchmarkFullTick(b *testing.B) {
	sc := &scenario.Scenario{
		ID:         "bench_scenario",
		Name:       "Benchmark",
		Difficulty: "easy",
		InitialState: scenario.InitialState{
			CaveWidth:      30,
			CaveHeight:     30,
			TerrainSeed:    42,
			TerrainDensity: 0.0,
			PrebuiltRooms: []scenario.RoomPlacement{
				{TypeID: "dragon_hole", Pos: types.Pos{X: 5, Y: 5}, Level: 1},
				{TypeID: "chi_chamber", Pos: types.Pos{X: 10, Y: 5}, Level: 1},
				{TypeID: "senju_room", Pos: types.Pos{X: 15, Y: 5}, Level: 1},
				{TypeID: "trap_room", Pos: types.Pos{X: 20, Y: 5}, Level: 1},
				{TypeID: "recovery_room", Pos: types.Pos{X: 5, Y: 10}, Level: 1},
			},
			DragonVeins: []scenario.DragonVeinPlacement{
				{SourcePos: types.Pos{X: 5, Y: 7}, Element: types.Wood, FlowRate: 10.0},
			},
			StartingChi: 500.0,
		},
		WinConditions: []scenario.ConditionDef{
			{Type: "survive_until", Params: map[string]any{"ticks": float64(200)}},
		},
		LoseConditions: []scenario.ConditionDef{
			{Type: "core_destroyed", Params: map[string]any{}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng := types.NewSeededRNG(int64(i))
		engine, err := NewSimulationEngine(sc, rng)
		if err != nil {
			b.Fatalf("NewSimulationEngine: %v", err)
		}
		noActions := []PlayerAction{NoAction{}}
		for tick := range 100 {
			_, err := engine.Step(noActions)
			if err != nil {
				b.Fatalf("Step tick %d: %v", tick, err)
			}
		}
	}
}
