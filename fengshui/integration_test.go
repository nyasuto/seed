package fengshui

import (
	"slices"
	"testing"

	"github.com/ponpoko/chaosseed-core/types"
	"github.com/ponpoko/chaosseed-core/world"
)

// TestIntegration_FengShuiSimulation verifies the full feng shui workflow on a
// 32x32 cave with 5 rooms (including generates and overcomes pairs), 2 dragon
// veins, and 20 ticks of chi flow simulation.
//
// Room layout:
//
//	Room 1 (Wood)  at (2,2)  3x3 — entrance south
//	Room 2 (Fire)  at (8,2)  3x3 — entrance south  (Wood→Fire = generates)
//	Room 3 (Earth) at (14,2) 3x3 — entrance south  (Wood→Earth = overcomes)
//	Room 4 (Metal) at (5,10) 3x3 — entrance north
//	Room 5 (Water) at (14,10) 3x3 — entrance north (Water→Wood = generates)
//
// Connections: 1-2, 2-3, 1-4, 4-5, 3-5 (cycle)
// Dragon vein 1: Wood element, from (3,5) — reaches rooms 1,2,3
// Dragon vein 2: Water element, from (6,9) — reaches rooms 4,5
func TestIntegration_FengShuiSimulation(t *testing.T) {
	cave, err := world.NewCave(32, 32)
	if err != nil {
		t.Fatalf("NewCave: %v", err)
	}

	reg := testRegistry()

	// --- Place 5 rooms ---

	type roomDef struct {
		typeID    string
		pos       types.Pos
		entrances []world.RoomEntrance
	}

	defs := []roomDef{
		{ // Room 1: Wood
			typeID: "wood_room",
			pos:    types.Pos{X: 2, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 3, Y: 4}, Dir: types.South},
			},
		},
		{ // Room 2: Fire (generates pair with Wood)
			typeID: "fire_room",
			pos:    types.Pos{X: 8, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 9, Y: 4}, Dir: types.South},
			},
		},
		{ // Room 3: Earth (overcomes pair with Wood — Wood overcomes Earth)
			typeID: "earth_room",
			pos:    types.Pos{X: 14, Y: 2},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 15, Y: 4}, Dir: types.South},
			},
		},
		{ // Room 4: Metal
			typeID: "metal_room",
			pos:    types.Pos{X: 5, Y: 10},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 6, Y: 9}, Dir: types.North},
			},
		},
		{ // Room 5: Water
			typeID: "water_room",
			pos:    types.Pos{X: 14, Y: 10},
			entrances: []world.RoomEntrance{
				{Pos: types.Pos{X: 15, Y: 9}, Dir: types.North},
			},
		},
	}

	rooms := make([]*world.Room, 0, len(defs))
	for i, d := range defs {
		r, err := cave.AddRoom(d.typeID, d.pos, 3, 3, d.entrances)
		if err != nil {
			t.Fatalf("AddRoom[%d] %s: %v", i, d.typeID, err)
		}
		rooms = append(rooms, r)
	}

	if len(cave.Rooms) != 5 {
		t.Fatalf("expected 5 rooms, got %d", len(cave.Rooms))
	}

	// --- Connect rooms ---
	// 1-2, 2-3, 1-4, 4-5, 3-5
	connections := [][2]int{
		{rooms[0].ID, rooms[1].ID},
		{rooms[1].ID, rooms[2].ID},
		{rooms[0].ID, rooms[3].ID},
		{rooms[3].ID, rooms[4].ID},
		{rooms[2].ID, rooms[4].ID},
	}
	for i, c := range connections {
		_, err := cave.ConnectRooms(c[0], c[1])
		if err != nil {
			t.Fatalf("ConnectRooms[%d] (%d,%d): %v", i, c[0], c[1], err)
		}
	}

	// Verify adjacency graph is fully connected.
	graph := cave.BuildAdjacencyGraph()
	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			if !graph.PathExists(rooms[i].ID, rooms[j].ID) {
				t.Fatalf("rooms %d and %d are not connected", rooms[i].ID, rooms[j].ID)
			}
		}
	}

	// --- Build 2 dragon veins ---
	// Vein 1: Wood element, sourced from the corridor area south of Room 1.
	// Low flow rate so rooms don't hit capacity and we can observe element affinity effects.
	vein1, err := BuildDragonVein(cave, types.Pos{X: 3, Y: 5}, types.Wood, 2.0)
	if err != nil {
		t.Fatalf("BuildDragonVein 1: %v", err)
	}

	// Vein 2: Water element, sourced from corridor area north of Room 4.
	vein2, err := BuildDragonVein(cave, types.Pos{X: 6, Y: 9}, types.Water, 2.0)
	if err != nil {
		t.Fatalf("BuildDragonVein 2: %v", err)
	}

	veins := []*DragonVein{vein1, vein2}

	// Verify dragon veins reach at least one room each.
	if len(vein1.RoomsOnPath(cave)) == 0 {
		t.Fatal("dragon vein 1 reaches no rooms")
	}
	if len(vein2.RoomsOnPath(cave)) == 0 {
		t.Fatal("dragon vein 2 reaches no rooms")
	}

	// --- Run simulation: 20 ticks ---
	flowParams := DefaultFlowParams()
	engine := NewChiFlowEngine(cave, veins, reg, flowParams)

	if len(engine.RoomChi) != 5 {
		t.Fatalf("RoomChi count = %d, want 5", len(engine.RoomChi))
	}

	for range 20 {
		engine.Tick()
	}

	// --- Verify chi levels ---
	// All rooms on dragon vein paths should have positive chi.
	for _, rc := range engine.RoomChi {
		if rc.Current < 0 {
			t.Errorf("room %d chi = %v, should not be negative", rc.RoomID, rc.Current)
		}
		if rc.Current > rc.Capacity {
			t.Errorf("room %d chi = %v exceeds capacity %v", rc.RoomID, rc.Current, rc.Capacity)
		}
	}

	// --- Verify generates vs overcomes chi effect ---
	// Dragon vein 1 is Wood element.
	// Room 1 (Wood): same element → 1.1x supply
	// Room 2 (Fire): Wood generates Fire → 1.3x supply (if on vein path)
	// Room 3 (Earth): Wood overcomes Earth → 0.6x supply (if on vein path)
	//
	// Rooms on vein 1 that are generates-paired should have more chi than
	// overcomes-paired rooms.

	vein1Rooms := vein1.RoomsOnPath(cave)
	room2OnVein1 := false
	room3OnVein1 := false
	for _, rid := range vein1Rooms {
		if rid == rooms[1].ID {
			room2OnVein1 = true
		}
		if rid == rooms[2].ID {
			room3OnVein1 = true
		}
	}

	if room2OnVein1 && room3OnVein1 {
		// Both rooms receive supply from the same Wood vein.
		// Fire room (generates) should accumulate more chi than Earth room (overcomes).
		chiRoom2 := engine.RoomChi[rooms[1].ID].Current
		chiRoom3 := engine.RoomChi[rooms[2].ID].Current
		if chiRoom2 <= chiRoom3 {
			t.Errorf("generates pair: Fire room chi (%v) should be > Earth room chi (%v) "+
				"when supplied by Wood dragon vein", chiRoom2, chiRoom3)
		}
	}

	// --- Verify feng shui scores ---
	scoreParams := DefaultScoreParams()
	evaluator := NewEvaluator(cave, reg, scoreParams)

	scores := evaluator.EvaluateAll(engine)
	if len(scores) != 5 {
		t.Fatalf("EvaluateAll returned %d scores, want 5", len(scores))
	}

	// All scores should have a non-negative Total (chi contributes positively,
	// adjacency can be negative but dragon vein bonus offsets it).
	for _, s := range scores {
		if s.ChiScore < 0 {
			t.Errorf("room %d ChiScore = %v, should be >= 0", s.RoomID, s.ChiScore)
		}
	}

	// Verify that rooms on dragon veins get the DragonVeinBonus.
	for _, s := range scores {
		onVein := false
		for _, vein := range veins {
			if slices.Contains(vein.RoomsOnPath(cave), s.RoomID) {
				onVein = true
			}
			if onVein {
				break
			}
		}
		if onVein && s.DragonVeinScore != scoreParams.DragonVeinBonus {
			t.Errorf("room %d on dragon vein: DragonVeinScore = %v, want %v",
				s.RoomID, s.DragonVeinScore, scoreParams.DragonVeinBonus)
		}
	}

	// Room 1 (Wood) is adjacent to Room 2 (Fire): Wood generates Fire → +20 bonus.
	// Room 1 (Wood) is adjacent to Room 4 (Metal): Metal overcomes Wood → -15 penalty
	//   (from Room 1's perspective: Overcomes(Wood, Metal)? No. Generates(Wood,Metal)? No.
	//    Overcomes(Metal,Wood)? Yes, but adjacencyBonus checks (roomElem, neighborElem),
	//    so from Room 1: Overcomes(Wood, Metal) = false, Generates(Wood, Metal) = false → 0)
	// Let's verify the adjacency score for room 1-2 pair.
	score1 := evaluator.EvaluateRoom(rooms[0].ID, engine)
	// Room 1 neighbors: Room 2 (Fire), Room 4 (Metal)
	// adjacencyBonus(Wood, Fire) → Generates(Wood, Fire) = true → +20
	// adjacencyBonus(Wood, Metal) → neither generates nor overcomes → 0
	expectedAdj1 := scoreParams.GeneratesBonus // +20 for Wood→Fire neighbor
	if score1.AdjacencyScore != expectedAdj1 {
		t.Errorf("room 1 AdjacencyScore = %v, want %v (Wood adj to Fire=generates, Metal=neutral)",
			score1.AdjacencyScore, expectedAdj1)
	}

	// Room 3 (Earth) neighbors: Room 2 (Fire), Room 5 (Water)
	// adjacencyBonus(Earth, Fire) → Generates(Earth, Fire)? No. Overcomes(Earth, Fire)? No. → 0
	//   Actually: Generates: Wood→Fire, Fire→Earth, Earth→Metal, Metal→Water, Water→Wood
	//   So Generates(Earth, Fire) = false. Overcomes: Wood→Earth, Fire→Metal, Earth→Water,
	//   Metal→Wood, Water→Fire. Overcomes(Earth, Fire) = false.
	//   But what about the reverse? We check (roomElem=Earth, neighborElem=Fire): neutral.
	// adjacencyBonus(Earth, Water) → Overcomes(Earth, Water) = true → -15
	score3 := evaluator.EvaluateRoom(rooms[2].ID, engine)
	expectedAdj3 := scoreParams.OvercomesPenalty // -15 for Earth→Water (overcomes)
	if score3.AdjacencyScore != expectedAdj3 {
		t.Errorf("room 3 AdjacencyScore = %v, want %v (Earth adj to Fire=neutral, Water=overcomes)",
			score3.AdjacencyScore, expectedAdj3)
	}

	// CaveTotal should equal sum of all individual scores.
	caveTotal := evaluator.CaveTotal(engine)
	var sumTotal float64
	for _, s := range scores {
		sumTotal += s.Total
	}
	if caveTotal != sumTotal {
		t.Errorf("CaveTotal = %v, want sum of scores = %v", caveTotal, sumTotal)
	}

	// CaveTotal should be positive overall (dragon vein bonuses and chi outweigh penalties).
	if caveTotal <= 0 {
		t.Errorf("CaveTotal = %v, expected positive for a well-connected cave", caveTotal)
	}
}
